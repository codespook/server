package server

type QuestionType string

const LIKERT QuestionType = "likert"

type Aggregation string

const (
	MEAN Aggregation = "mean"
	SUM  Aggregation = "sum"
)

type Question struct {
	ID         string                 `json:"id"`
	Question   string                 `json:"question"`
	Type       QuestionType           `json:"type"`
	Deleted    bool                   `json:"deleted"`
	Options    map[string]interface{} `json:"options"`
	CategoryID string                 `json:"categoryID"  bson:"categoryID"`
}

type Category struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Aggregation Aggregation `json:"aggregation"`
}

type OutcomeSet struct {
	ID             string     `json:"id" bson:"_id"`
	OrganisationID string     `json:"organisationID" bson:"organisationID"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Questions      []Question `json:"questions"`
	Categories     []Category `json:"categories"`
	Deleted        bool       `json:"deleted"`
}

func (os *OutcomeSet) GetCategory(catID string) *Category {
	for _, c := range os.Categories {
		if c.ID == catID {
			return &c
		}
	}
	return nil
}

// GetCategoryQuestions gets questions belonging to the provided category ID
// Does not return archived questions
func (os *OutcomeSet) GetCategoryQuestions(catID string) []Question {
	out := make([]Question, 0, len(os.Questions))
	for _, q := range os.Questions {
		if q.CategoryID == catID && !q.Deleted {
			out = append(out, q)
		}
	}
	return out
}

// GetArchivedCategoryQuestions gets archived questions belonging to the provided category ID
func (os *OutcomeSet) GetArchivedCategoryQuestions(catID string) []Question {
	out := make([]Question, 0, len(os.Questions))
	for _, q := range os.Questions {
		if q.CategoryID == catID && q.Deleted {
			out = append(out, q)
		}
	}
	return out
}

// GetQuestion returns the question with the provided qID or nil
func (os *OutcomeSet) GetQuestion(qID string) *Question {
	for _, q := range os.Questions {
		if q.ID == qID {
			return &q
		}
	}
	return nil
}

// ActiveQuestions returns only the currently active questions (i.e. not deleted)
func (os *OutcomeSet) ActiveQuestions() []Question {
	qs := make([]Question, 0, len(os.Questions))
	for _, q := range os.Questions {
		if !q.Deleted {
			qs = append(qs, q)
		}
	}
	return qs
}
