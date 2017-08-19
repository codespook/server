package logic

import (
	"fmt"
	impact "github.com/impactasaurus/server"
	"github.com/impactasaurus/server/auth"
	"github.com/impactasaurus/server/data"
	"log"
	"time"
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
		// 	 find first meeting
		var firstMeeting *impact.Meeting
		for _, meeting := range benMeetings {
			if firstMeeting == nil || firstMeeting.Conducted.After(meeting.Conducted) {
				firstMeeting = &meeting
			}
		}
		if firstMeeting == nil {
			j.addGlobalWarning(fmt.Sprintf("Beneficiary %s was not included as they only have a single meeting recorded", ben))
			continue
		}
		firstAndLast[ben] = firstAndLastMeetings{
			first: *firstMeeting,
			last:  lastMeeting,
		}
	}
	return firstAndLast
}

func (j *jocReporter) getQuestionAggregations(firstAndLast map[string]firstAndLastMeetings) impact.JOCQAggs {
	return impact.JOCQAggs{}
}

func (j *jocReporter) getCategoryAggregations(firstAndLast map[string]firstAndLastMeetings) impact.JOCCatAggs {
	return impact.JOCCatAggs{}
}

func (j *jocReporter) getBeneficiaryIDs(firstAndLast map[string]firstAndLastMeetings) []string {
	bens := make([]string, 0, len(firstAndLast))
	for b := range firstAndLast {
		bens = append(bens, b)
	}
	return bens
}

func GetJOCServiceReport(start, end time.Time, questionSetID string, db data.Base, u auth.User) (*impact.JOCServiceReport, error) {
	os, err := db.GetOutcomeSet(questionSetID, u)
	if err != nil {
		return nil, err
	}
	j := jocReporter{
		questionSetID: questionSetID,
		db:            db,
		u:             u,
		os:            os,
	}
	// DB: get meetings in time range for a question set
	meetingsInRange, err := db.GetOSMeetingsInTimeRange(start, end, questionSetID, u)
	if err != nil {
		return nil, err
	}
	// find last meeting for each beneficiary
	lastMeetings := j.getLastMeetingForEachBen(meetingsInRange)
	// for each ben
	firstAndLast := j.getFirstAndLastMeetings(lastMeetings)
	//   check both are full
	// 	 calc delta
	//   store
	// aggregate
	qAggs := j.getQuestionAggregations(firstAndLast)
	cAggs := j.getCategoryAggregations(firstAndLast)

	return &impact.JOCServiceReport{
		Excluded: impact.Excluded{
			CategoryIDs: j.excludedCategoryIDs,
			QuestionIDs: j.excludedQuestionIDs,
		},
		BeneficiaryIDs:     j.getBeneficiaryIDs(firstAndLast),
		CategoryAggregates: cAggs,
		QuestionAggregates: qAggs,
		Warnings:           j.globalWarnings,
	}, nil
}
