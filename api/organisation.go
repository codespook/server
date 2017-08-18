package api

import (
	"github.com/graphql-go/graphql"
)

func (v *v1) initOrgTypes() organisationTypes {
	return organisationTypes{
		organisationType: graphql.NewObject(graphql.ObjectConfig{
			Name:        "Organisation",
			Description: "An organisation",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type:        graphql.NewNonNull(graphql.ID),
					Description: "Unique identifier for the organisation",
				},
				"name": &graphql.Field{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "Organisation's name",
				},
			},
		}),
	}
}
