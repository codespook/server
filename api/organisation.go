package api

import (
	"github.com/graphql-go/graphql"
	"github.com/impactasaurus/server/auth"
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

func (v *v1) getOrgQueries(orgTypes organisationTypes) graphql.Fields {
	return graphql.Fields{
		"organisation": &graphql.Field{
			Type:        orgTypes.organisationType,
			Description: "Get an organisation by ID",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Description: "The ID of the organisation",
					Type:        graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
				return v.db.GetOrganisation(p.Args["id"].(string), u)
			}),
		},
	}
}
