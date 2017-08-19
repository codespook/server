package api

import (
	"fmt"
	"github.com/graphql-go/graphql"
)

func combineFields(toCombine ...graphql.Fields) (graphql.Fields, error) {
	final := graphql.Fields{}
	for _, toAdd := range toCombine {
		for k, v := range toAdd {
			if _, ok := final[k]; ok {
				return nil, fmt.Errorf("Query with name %s already exists", k)
			}
			final[k] = v
		}
	}
	return final, nil
}

func (v *v1) getSchema(orgTypes organisationTypes, osTypes outcomeSetTypes, meetTypes meetingTypes) (*graphql.Schema, error) {
	queries, err := combineFields(
		v.getMeetingQueries(meetTypes),
		v.getOrgQueries(orgTypes),
		v.getOSQueries(osTypes),
	)
	if err != nil {
		return nil, err
	}
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: queries,
	})

	mutations, err := combineFields(
		v.getOSMutations(osTypes),
		v.getMeetingMutations(meetTypes),
	)

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Mutation",
		Fields: mutations,
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
		Types: []graphql.Type{
			osTypes.likertScale,
			meetTypes.intAnswer,
		},
	})
	if err != nil {
		return nil, err
	}
	return &schema, nil
}
