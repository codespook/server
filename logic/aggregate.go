package logic

import (
	"errors"
	"fmt"
	impact "github.com/impactasaurus/server"
)

func mean(in []float32) float32 {
	s := sum(in)
	return s / float32(len(in))
}

func sum(in []float32) float32 {
	var total float32 = 0
	for _, value := range in {
		total += value
	}
	return total
}

func aggregate(in []float32, aggregation impact.Aggregation) (float32, error) {
	switch aggregation {
	case impact.MEAN:
		return mean(in), nil
	case impact.SUM:
		return sum(in), nil
	default:
		return 0, errors.New("Unknown aggregation")
	}
}

// GetCategoryAggregate aggregates multiple answers into a single value.
// If the returned CategoryAggregate is nil, there were no answers available for the category.
func GetCategoryAggregate(m impact.Meeting, categoryID string, os impact.OutcomeSet) (*impact.CategoryAggregate, error) {
	c := os.GetCategory(categoryID)
	if c == nil {
		return nil, fmt.Errorf("Couldn't find category %s", categoryID)
	}
	vals := make([]float32, 0, len(m.Answers))
	for _, a := range m.Answers {
		q := os.GetQuestion(a.QuestionID)
		if q.CategoryID == categoryID {
			f, err := a.ToFloat()
			if err != nil {
				return nil, err
			}
			vals = append(vals, f)
		}
	}
	if len(vals) == 0 {
		return nil, nil
	}
	ag, err := aggregate(vals, c.Aggregation)
	if err != nil {
		return nil, err
	}
	return &impact.CategoryAggregate{
		CategoryID: c.ID,
		Value:      ag,
	}, nil
}

func GetCategoryAggregates(m impact.Meeting, os impact.OutcomeSet) ([]impact.CategoryAggregate, error) {
	out := make([]impact.CategoryAggregate, 0, len(os.Categories))
	for _, c := range os.Categories {
		catAg, err := GetCategoryAggregate(m, c.ID, os)
		if err != nil {
			return nil, err
		}
		if catAg != nil {
			out = append(out, *catAg)
		}
	}
	return out, nil
}
