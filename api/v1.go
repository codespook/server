package api

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/impactasaurus/server/data"
	"net/http"
)

type meetingsTypes struct {
	answerInterface   *graphql.Interface
	intAnswer         *graphql.Object
	categoryAggregate *graphql.Object
	aggregates        *graphql.Object
	meetingType       *graphql.Object
}

type organisationTypes struct {
	organisationType *graphql.Object
}

type outcomeSetTypes struct {
	questionInterface *graphql.Interface
	likertScale       *graphql.Object
	outcomeSetType    *graphql.Object
	aggregationEnum   *graphql.Enum
	categoryType      *graphql.Object
}

type v1 struct {
	db data.Base

	meetingsTypes
	organisationTypes
	outcomeSetTypes
}

func NewV1(db data.Base) (http.Handler, error) {
	v := &v1{
		db: db,
	}
	v.initSchemaTypes()
	schema, err := v.getSchema()
	if err != nil {
		return nil, err
	}
	h := handler.New(&handler.Config{
		Schema: schema,
		Pretty: true,
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ContextHandler(r.Context(), w, r)
	}), nil
}
