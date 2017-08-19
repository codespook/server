package data

import (
	"fmt"
	impact "github.com/impactasaurus/server"
	"github.com/impactasaurus/server/auth"
	"time"
)

// NotFound is an error struct to be used when request to fetch / update / delete an individual element fails because the element does not exist.
type notFound struct {
	thing string
}

func NewNotFoundError(thing string) error {
	return &notFound{thing}
}

func (nf *notFound) Error() string {
	return fmt.Sprintf("%s not found", nf.thing)
}

type Base interface {
	NewOutcomeSet(name, description string, u auth.User) (impact.OutcomeSet, error)
	EditOutcomeSet(id, name, description string, u auth.User) (impact.OutcomeSet, error)
	GetOutcomeSet(id string, u auth.User) (impact.OutcomeSet, error)
	GetOutcomeSets(u auth.User) ([]impact.OutcomeSet, error)
	DeleteOutcomeSet(id string, u auth.User) error

	GetQuestion(outcomeSetID string, questionID string, u auth.User) (impact.Question, error)
	NewQuestion(outcomeSetID, question string, questionType impact.QuestionType, options map[string]interface{}, u auth.User) (impact.Question, error)
	DeleteQuestion(outcomeSetID, questionID string, u auth.User) error
	EditQuestion(outcomeSetID, questionID, question string, questionType impact.QuestionType, options map[string]interface{}, u auth.User) (impact.Question, error)

	GetCategory(outcomeSetID, categoryID string, u auth.User) (impact.Category, error)
	NewCategory(outcomeSetID, name, description string, aggregation impact.Aggregation, u auth.User) (impact.Category, error)
	DeleteCategory(outcomeSetID, categoryID string, u auth.User) error
	SetCategory(outcomeSetID, questionID, categoryID string, u auth.User) (impact.Question, error)
	RemoveCategory(outcomeSetID, questionID string, u auth.User) (impact.Question, error)

	GetOrganisation(id string, u auth.User) (impact.Organisation, error)

	GetMeeting(id string, u auth.User) (impact.Meeting, error)
	GetMeetingsForBeneficiary(beneficiary string, u auth.User) ([]impact.Meeting, error)
	GetOSMeetingsInTimeRange(start, end time.Time, outcomeSetID string, u auth.User) ([]impact.Meeting, error)
	GetOSMeetingsForBeneficiary(beneficiary string, outcomeSetID string, u auth.User) ([]impact.Meeting, error)
	NewMeeting(beneficiaryID, outcomeSetID string, conducted time.Time, u auth.User) (impact.Meeting, error)
	NewAnswer(meetingID string, answer impact.Answer, u auth.User) (impact.Meeting, error)
}
