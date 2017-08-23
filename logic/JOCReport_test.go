package logic_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	impact "github.com/impactasaurus/server"
	"github.com/impactasaurus/server/logic"
	"github.com/impactasaurus/server/mock"
	"github.com/stretchr/testify/assert"
)

func setupWrapper(t *testing.T, inner func(*mock.MockUser, *mock.MockBase)) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockUser := mock.NewMockUser(mockCtrl)
	mockDB := mock.NewMockBase(mockCtrl)
	inner(mockUser, mockDB)
}

func getDefaultMeetings(start, end time.Time, questionSetID string) map[string]impact.Meeting {
	return map[string]impact.Meeting{
		"B1M1": {
			ID:           "B1M1",
			Beneficiary:  "B1",
			OutcomeSetID: questionSetID,
			Conducted:    start.Add(-time.Hour * 84),
			Answers: []impact.Answer{{
				QuestionID: "Q1",
				Type:       impact.INT,
				Answer:     5,
			}, {
				QuestionID: "Q2",
				Type:       impact.INT,
				Answer:     5,
			}, {
				QuestionID: "Q3",
				Type:       impact.INT,
				Answer:     5,
			}, {
				QuestionID: "Q4",
				Type:       impact.INT,
				Answer:     5,
			}},
		},
		"B1M2": {
			ID:           "B1M2",
			Beneficiary:  "B1",
			OutcomeSetID: questionSetID,
			Conducted:    end,
			Answers: []impact.Answer{{
				QuestionID: "Q1",
				Type:       impact.INT,
				Answer:     9,
			}, {
				QuestionID: "Q2",
				Type:       impact.INT,
				Answer:     8,
			}, {
				QuestionID: "Q3",
				Type:       impact.INT,
				Answer:     8,
			}, {
				QuestionID: "Q4",
				Type:       impact.INT,
				Answer:     5,
			}},
		},
		"B2M1": {
			ID:           "B2M1",
			Beneficiary:  "B2",
			OutcomeSetID: questionSetID,
			Conducted:    start.Add(time.Hour),
			Answers: []impact.Answer{{
				QuestionID: "Q1",
				Type:       impact.INT,
				Answer:     6,
			}, {
				QuestionID: "Q2",
				Type:       impact.INT,
				Answer:     2,
			}, {
				QuestionID: "Q3",
				Type:       impact.INT,
				Answer:     7,
			}, {
				QuestionID: "Q4",
				Type:       impact.INT,
				Answer:     4,
			}},
		},
		"B2M2": {
			ID:           "B2M2",
			Beneficiary:  "B2",
			OutcomeSetID: questionSetID,
			Conducted:    end,
			Answers: []impact.Answer{{
				QuestionID: "Q1",
				Type:       impact.INT,
				Answer:     2,
			}, {
				QuestionID: "Q2",
				Type:       impact.INT,
				Answer:     2,
			}, {
				QuestionID: "Q3",
				Type:       impact.INT,
				Answer:     3,
			}, {
				QuestionID: "Q4",
				Type:       impact.INT,
				Answer:     5,
			}},
		},
		"B3M1": {
			ID:           "B3M1",
			Beneficiary:  "B3",
			OutcomeSetID: questionSetID,
			Conducted:    start.Add(-time.Hour),
			Answers: []impact.Answer{{
				QuestionID: "Q1",
				Type:       impact.INT,
				Answer:     1,
			}, {
				QuestionID: "Q2",
				Type:       impact.INT,
				Answer:     2,
			}, {
				QuestionID: "Q3",
				Type:       impact.INT,
				Answer:     3,
			}, {
				QuestionID: "Q4",
				Type:       impact.INT,
				Answer:     4,
			}},
		},
		"B3M2": {
			ID:           "B3M2",
			Beneficiary:  "B3",
			OutcomeSetID: questionSetID,
			Conducted:    start.Add(time.Hour),
			Answers: []impact.Answer{{
				QuestionID: "Q1",
				Type:       impact.INT,
				Answer:     10,
			}, {
				QuestionID: "Q2",
				Type:       impact.INT,
				Answer:     10,
			}, {
				QuestionID: "Q3",
				Type:       impact.INT,
				Answer:     10,
			}, {
				QuestionID: "Q4",
				Type:       impact.INT,
				Answer:     10,
			}},
		},
		"B3M3": {
			ID:           "B3M3",
			Beneficiary:  "B3",
			OutcomeSetID: questionSetID,
			Conducted:    end,
			Answers: []impact.Answer{{
				QuestionID: "Q1",
				Type:       impact.INT,
				Answer:     5,
			}, {
				QuestionID: "Q2",
				Type:       impact.INT,
				Answer:     5,
			}, {
				QuestionID: "Q3",
				Type:       impact.INT,
				Answer:     5,
			}, {
				QuestionID: "Q4",
				Type:       impact.INT,
				Answer:     6,
			}},
		},
	}
}

func TestJOCReport(t *testing.T) {
	questionSetID := "qid"
	end := time.Now()
	start := end.Add(-time.Hour * 24)

	os := impact.OutcomeSet{
		ID: questionSetID,
		Questions: []impact.Question{{
			ID:         "Q1",
			Type:       impact.LIKERT,
			CategoryID: "C1",
		}, {
			ID:         "Q2",
			Type:       impact.LIKERT,
			CategoryID: "C1",
		}, {
			ID:         "Q3",
			Type:       impact.LIKERT,
			CategoryID: "C2",
		}, {
			ID:         "Q4",
			Type:       impact.LIKERT,
			CategoryID: "C2",
		}},
		Categories: []impact.Category{{
			ID:          "C1",
			Aggregation: impact.MEAN,
		}, {
			ID:          "C2",
			Aggregation: impact.MEAN,
		}},
	}

	meetings := getDefaultMeetings(start, end, questionSetID)

	inRangeMeetings := []impact.Meeting{
		meetings["B1M2"],
		meetings["B2M1"],
		meetings["B2M2"],
		meetings["B3M2"],
		meetings["B3M3"],
	}

	b1Meetings := []impact.Meeting{meetings["B1M1"], meetings["B1M2"]}
	b2Meetings := []impact.Meeting{meetings["B2M1"], meetings["B2M2"]}
	b3Meetings := []impact.Meeting{meetings["B3M1"], meetings["B3M2"], meetings["B3M3"]}

	expected := impact.JOCServiceReport{
		BeneficiaryIDs: []string{"B1", "B2", "B3"},
		Warnings:       []string{},
		Excluded: impact.Excluded{
			CategoryIDs: []string{},
			QuestionIDs: []string{},
		},
		QuestionAggregates: impact.JOCQAggs{
			First: []impact.QBenAgg{{
				QuestionID:     "Q1",
				Value:          4,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}, {
				QuestionID:     "Q2",
				Value:          3,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}, {
				QuestionID:     "Q3",
				Value:          5,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}, {
				QuestionID:     "Q4",
				Value:          4.3333335,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}},
			Last: []impact.QBenAgg{{
				QuestionID:     "Q1",
				Value:          5.3333335,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}, {
				QuestionID:     "Q2",
				Value:          5,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}, {
				QuestionID:     "Q3",
				Value:          5.3333335,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}, {
				QuestionID:     "Q4",
				Value:          5.3333335,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}},
			Delta: []impact.QBenAgg{{
				QuestionID:     "Q1",
				Value:          1.3333334,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}, {
				QuestionID:     "Q2",
				Value:          2,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}, {
				QuestionID:     "Q3",
				Value:          0.33333334,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}, {
				QuestionID:     "Q4",
				Value:          1,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}},
		},
		CategoryAggregates: impact.JOCCatAggs{
			First: []impact.CatBenAgg{{
				CategoryID:     "C1",
				Value:          3.5,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}, {
				CategoryID:     "C2",
				Value:          4.6666665,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}},
			Last: []impact.CatBenAgg{{
				CategoryID:     "C1",
				Value:          5.1666665,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}, {
				CategoryID:     "C2",
				Value:          5.3333335,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}},
			Delta: []impact.CatBenAgg{{
				CategoryID:     "C1",
				Value:          1.6666666,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}, {
				CategoryID:     "C2",
				Value:          0.6666667,
				BeneficiaryIDs: []string{"B1", "B2", "B3"},
				Warnings:       []string{},
			}},
		},
	}

	setupWrapper(t, func(mockUser *mock.MockUser, mockDB *mock.MockBase) {
		mockDB.EXPECT().GetOutcomeSet(questionSetID, mockUser).Return(os, nil)
		mockDB.EXPECT().GetOSMeetingsInTimeRange(start, end, questionSetID, mockUser).Return(inRangeMeetings, nil)
		mockDB.EXPECT().GetOSMeetingsForBeneficiary("B1", questionSetID, mockUser).Return(b1Meetings, nil)
		mockDB.EXPECT().GetOSMeetingsForBeneficiary("B2", questionSetID, mockUser).Return(b2Meetings, nil)
		mockDB.EXPECT().GetOSMeetingsForBeneficiary("B3", questionSetID, mockUser).Return(b3Meetings, nil)

		result, err := logic.GetJOCServiceReport(start, end, questionSetID, mockDB, mockUser)
		assert.NoError(t, err)
		assert.EqualValues(t, expected, *result)
	})

}

func TestOutcomeSetError(t *testing.T) {
	setupWrapper(t, func(mockUser *mock.MockUser, mockDB *mock.MockBase) {
		e := errors.New("Mongo error")
		mockDB.EXPECT().GetOutcomeSet("q", mockUser).Return(impact.OutcomeSet{}, e)
		result, err := logic.GetJOCServiceReport(time.Now(), time.Now(), "q", mockDB, mockUser)
		assert.Nil(t, result)
		assert.EqualError(t, err, e.Error())
	})
}

func TestMeetingsInRangeError(t *testing.T) {
	setupWrapper(t, func(mockUser *mock.MockUser, mockDB *mock.MockBase) {
		e := errors.New("Mongo error")
		mockDB.EXPECT().GetOutcomeSet("q", mockUser).Return(impact.OutcomeSet{}, nil)
		mockDB.EXPECT().GetOSMeetingsInTimeRange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, e)
		result, err := logic.GetJOCServiceReport(time.Now(), time.Now(), "q", mockDB, mockUser)
		assert.Nil(t, result)
		assert.EqualError(t, err, e.Error())
	})
}

func TestNoMeetingsInRange(t *testing.T) {
	meetingsInRange := []impact.Meeting{}

	setupWrapper(t, func(mockUser *mock.MockUser, mockDB *mock.MockBase) {
		mockDB.EXPECT().GetOutcomeSet("q", mockUser).Return(impact.OutcomeSet{}, nil)
		mockDB.EXPECT().GetOSMeetingsInTimeRange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(meetingsInRange, nil)
		result, err := logic.GetJOCServiceReport(time.Now(), time.Now(), "q", mockDB, mockUser)
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestUsersWithOnlyOneMeeting(t *testing.T) {

	meetings := getDefaultMeetings(time.Now().Add(-time.Hour*10), time.Now(), "qid")
	meetingsInRange := []impact.Meeting{meetings["B1M1"]}

	setupWrapper(t, func(mockUser *mock.MockUser, mockDB *mock.MockBase) {
		mockDB.EXPECT().GetOutcomeSet("q", mockUser).Return(impact.OutcomeSet{}, nil)
		mockDB.EXPECT().GetOSMeetingsInTimeRange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(meetingsInRange, nil)
		mockDB.EXPECT().GetOSMeetingsForBeneficiary("B1", gomock.Any(), mockUser).Return(meetingsInRange, nil)
		result, err := logic.GetJOCServiceReport(time.Now(), time.Now(), "q", mockDB, mockUser)
		assert.NoError(t, err)
		assert.Len(t, result.BeneficiaryIDs, 0)
		assert.Len(t, result.Warnings, 1)
		assert.Regexp(t, regexp.MustCompile("Beneficiary .* only have a single meeting recorded"), result.Warnings[0])
	})
}

// categories with no questions
// questions with no questions
// beneficiaries missing first/last question
// getting ben's meeting failure
