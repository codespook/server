package server

import (
	"errors"
	"time"
)

type AnswerType string

const INT AnswerType = "int"

type Answer struct {
	QuestionID string      `json:"questionID" bson:"questionID"`
	Answer     interface{} `json:"answer"`
	Type       AnswerType  `json:"type"`
}

type Meeting struct {
	ID             string    `json:"id" bson:"_id"`
	Beneficiary    string    `json:"beneficiary"`
	User           string    `json:"user"`
	OutcomeSetID   string    `json:"outcomeSetID" bson:"outcomeSetID"`
	OrganisationID string    `json:"organisationID" bson:"organisationID"`
	Answers        []Answer  `json:"answers"`
	Conducted      time.Time `json:"conducted"`
	Created        time.Time `json:"created"`
	Modified       time.Time `json:"modified"`
}

// CategoryAggregate aggregates multiple questions belonging to the same category to a question category level
type CategoryAggregate struct {
	CategoryID string  `json:"categoryID"`
	Value      float32 `json:"value"`
}

// Aggregates stores aggregations associated with a meeting
type Aggregates struct {
	Category []CategoryAggregate `json:"category"`
}

func (a Answer) IsNumeric() bool {
	return a.Type == INT
}

func (a Answer) ToFloat() (float32, error) {
	switch i := a.Answer.(type) {
	case float32:
		return i, nil
	case float64:
		return float32(i), nil
	case int64:
		return float32(i), nil
	case int32:
		return float32(i), nil
	case int:
		return float32(i), nil
	default:
		return 0, errors.New("Cannot convert answer to float")
	}
}
