package logic

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	impact "github.com/impactasaurus/server"
	"github.com/impactasaurus/server/auth"
	"github.com/impactasaurus/server/data"
)

type firstAndLastMeetings struct {
	first impact.Meeting
	last  impact.Meeting
}

type jocReporter struct {
	questionSetID       string
	db                  data.Base
	u                   auth.User
	globalWarnings      []string
	os                  impact.OutcomeSet
	excludedCategoryIDs []string
	excludedQuestionIDs []string
}

func (j *jocReporter) addGlobalWarning(warning string) {
	j.globalWarnings = append(j.globalWarnings, warning)
}

func (j *jocReporter) getLastMeetingForEachBen(meetingsInRange []impact.Meeting) map[string]impact.Meeting {
	lastMeetings := map[string]impact.Meeting{}
	for _, meeting := range meetingsInRange {
		ben := meeting.Beneficiary
		existing, exists := lastMeetings[ben]
		record := !exists
		if exists && existing.Conducted.Before(meeting.Conducted) {
			record = true
		}
		if record {
			lastMeetings[ben] = meeting
		}
	}
	return lastMeetings
}

func (j *jocReporter) getFirstAndLastMeetings(lastMeetings map[string]impact.Meeting) map[string]firstAndLastMeetings {
	firstAndLast := map[string]firstAndLastMeetings{}
	for ben, lastMeeting := range lastMeetings {
		// 	 DB: get meetings for os
		benMeetings, err := j.db.GetOSMeetingsForBeneficiary(ben, j.questionSetID, j.u)
		if err != nil {
			j.addGlobalWarning(fmt.Sprintf("Could not include beneficiary %s due to an system error. Please contact support.", ben))
			log.Printf("Getting benificary's (%s) meetings failed: %s", ben, err.Error())
			continue
		}
		if len(benMeetings) == 0 {
			j.addGlobalWarning(fmt.Sprintf("Could not include beneficiary %s as we could not find their first meeting. Please contact support.", ben))
			continue
		}
		// 	 find first meeting
		var firstMeeting impact.Meeting
		found := false
		for _, meeting := range benMeetings {
			if meeting.ID != lastMeeting.ID &&
				(!found || firstMeeting.Conducted.After(meeting.Conducted)) {
				firstMeeting = meeting
				found = true
			}
		}
		if !found {
			j.addGlobalWarning(fmt.Sprintf("Beneficiary %s was not included as they only have a single meeting recorded", ben))
			continue
		}
		firstAndLast[ben] = firstAndLastMeetings{
			first: firstMeeting,
			last:  lastMeeting,
		}
	}
	return firstAndLast
}

type beneficiaryAggregation struct {
	first         []float32
	last          []float32
	diff          []float32
	beneficiaries []string
	warnings      []string
	aggTarget     string
}

func newBenAgg(aggTargetID string, noBens int) *beneficiaryAggregation {
	return &beneficiaryAggregation{
		first:         make([]float32, 0, noBens),
		last:          make([]float32, 0, noBens),
		diff:          make([]float32, 0, noBens),
		beneficiaries: make([]string, 0, noBens),
		warnings:      make([]string, 0, noBens),
		aggTarget:     aggTargetID,
	}
}

func (ba *beneficiaryAggregation) addBenificaryValues(benID string, first, last float32) {
	ba.beneficiaries = append(ba.beneficiaries, benID)
	sort.Strings(ba.beneficiaries)
	ba.first = append(ba.first, first)
	ba.last = append(ba.last, last)
	ba.diff = append(ba.diff, last-first)
}

func (ba *beneficiaryAggregation) addBenificaryWarning(warning string) {
	ba.warnings = append(ba.warnings, warning)
}

func (ba *beneficiaryAggregation) aggregateQuestions(j *jocReporter, aggs *impact.JOCQAggs) {
	if len(ba.first) == 0 {
		j.excludedQuestionIDs = append(j.excludedQuestionIDs, ba.aggTarget)
		return
	}
	getBenAgg := func(toAdd []float32) impact.QBenAgg {
		return impact.QBenAgg{
			QuestionID:     ba.aggTarget,
			Warnings:       ba.warnings,
			BeneficiaryIDs: ba.beneficiaries,
			Value:          mean(toAdd),
		}
	}
	aggs.First = append(aggs.First, getBenAgg(ba.first))
	aggs.Last = append(aggs.Last, getBenAgg(ba.last))
	aggs.Delta = append(aggs.Delta, getBenAgg(ba.diff))
}

func (ba *beneficiaryAggregation) aggregateCategories(j *jocReporter, aggs *impact.JOCCatAggs) {
	if len(ba.first) == 0 {
		j.excludedCategoryIDs = append(j.excludedCategoryIDs, ba.aggTarget)
		return
	}
	getBenAgg := func(toAdd []float32) impact.CatBenAgg {
		return impact.CatBenAgg{
			CategoryID:     ba.aggTarget,
			Warnings:       ba.warnings,
			BeneficiaryIDs: ba.beneficiaries,
			Value:          mean(toAdd),
		}
	}
	aggs.First = append(aggs.First, getBenAgg(ba.first))
	aggs.Last = append(aggs.Last, getBenAgg(ba.last))
	aggs.Delta = append(aggs.Delta, getBenAgg(ba.diff))
}

func (j *jocReporter) getQuestionAggregations(firstAndLast map[string]firstAndLastMeetings) impact.JOCQAggs {
	activeQs := j.os.ActiveQuestions()
	ret := impact.JOCQAggs{
		First: make([]impact.QBenAgg, 0, len(activeQs)),
		Last:  make([]impact.QBenAgg, 0, len(activeQs)),
		Delta: make([]impact.QBenAgg, 0, len(activeQs)),
	}
	for _, q := range activeQs {
		benAggregator := newBenAgg(q.ID, len(firstAndLast))
		for ben, fl := range firstAndLast {
			firstAnswer := fl.first.GetAnswer(q.ID)
			lastAnswer := fl.last.GetAnswer(q.ID)
			if firstAnswer == nil || lastAnswer == nil {
				benAggregator.addBenificaryWarning(fmt.Sprintf("Beneficiary %s not included as the question was not answered in both the first and last meetings", ben))
				continue
			}
			if !firstAnswer.IsNumeric() || !lastAnswer.IsNumeric() {
				benAggregator.addBenificaryWarning(fmt.Sprintf("Beneficiary %s not included as the answers were not of an expected format", ben))
				continue
			}
			fV, fE := firstAnswer.ToFloat()
			lV, lE := lastAnswer.ToFloat()
			if fE != nil || lE != nil {
				benAggregator.addBenificaryWarning(fmt.Sprintf("Beneficiary %s not included as the answers were not of an expected format", ben))
				continue
			}
			benAggregator.addBenificaryValues(ben, fV, lV)
		}
		benAggregator.aggregateQuestions(j, &ret)
	}
	return ret
}

func (j *jocReporter) getCategoryAggregations(firstAndLast map[string]firstAndLastMeetings) impact.JOCCatAggs {
	ret := impact.JOCCatAggs{
		First: make([]impact.CatBenAgg, 0, len(j.os.Categories)),
		Last:  make([]impact.CatBenAgg, 0, len(j.os.Categories)),
		Delta: make([]impact.CatBenAgg, 0, len(j.os.Categories)),
	}
	for _, cat := range j.os.Categories {
		benAggregator := newBenAgg(cat.ID, len(firstAndLast))
		for ben, fl := range firstAndLast {
			fCat, fE := GetCategoryAggregate(fl.first, cat.ID, j.os)
			sCat, sE := GetCategoryAggregate(fl.last, cat.ID, j.os)
			if fE != nil || sE != nil {
				benAggregator.addBenificaryWarning(fmt.Sprintf("Beneficiary %s not included because the category aggregation failed", ben))
				continue
			}
			if fCat == nil || sCat == nil {
				benAggregator.addBenificaryWarning(fmt.Sprintf("Beneficiary %s not included as they had no answers belonging to the category", ben))
				continue
			}
			benAggregator.addBenificaryValues(ben, fCat.Value, sCat.Value)
		}
		benAggregator.aggregateCategories(j, &ret)
	}
	return ret
}

func (j *jocReporter) getBeneficiaryIDs(firstAndLast map[string]firstAndLastMeetings) []string {
	bens := make([]string, 0, len(firstAndLast))
	for b := range firstAndLast {
		bens = append(bens, b)
	}
	sort.Strings(bens)
	return bens
}

func GetJOCServiceReport(start, end time.Time, questionSetID string, db data.Base, u auth.User) (*impact.JOCServiceReport, error) {
	os, err := db.GetOutcomeSet(questionSetID, u)
	if err != nil {
		return nil, err
	}
	j := jocReporter{
		questionSetID:       questionSetID,
		db:                  db,
		u:                   u,
		os:                  os,
		globalWarnings:      []string{},
		excludedCategoryIDs: []string{},
		excludedQuestionIDs: []string{},
	}

	meetingsInRange, err := db.GetOSMeetingsInTimeRange(start, end, questionSetID, u)
	if err != nil {
		return nil, err
	}
	if len(meetingsInRange) == 0 {
		return nil, errors.New("No meetings found for the question set within the given date range")
	}

	lastMeetings := j.getLastMeetingForEachBen(meetingsInRange)
	firstAndLast := j.getFirstAndLastMeetings(lastMeetings)
	qAggs := j.getQuestionAggregations(firstAndLast)
	cAggs := j.getCategoryAggregations(firstAndLast)

	ret := impact.JOCServiceReport{
		Excluded: impact.Excluded{
			CategoryIDs: j.excludedCategoryIDs,
			QuestionIDs: j.excludedQuestionIDs,
		},
		BeneficiaryIDs:     j.getBeneficiaryIDs(firstAndLast),
		CategoryAggregates: cAggs,
		QuestionAggregates: qAggs,
		Warnings:           j.globalWarnings,
	}
	return &ret, nil
}
