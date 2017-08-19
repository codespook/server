package api

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/impactasaurus/server/auth"
	"github.com/impactasaurus/server/logic"
	"strings"
	"time"
)

func (v *v1) initRepTypes() reportTypes {

	excluded := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Excluded",
		Description: "Details the questions or categories excluded from an aggregation",
		Fields: graphql.Fields{
			"categoryIDs": &graphql.Field{
				Type:        graphql.NewList(graphql.String),
				Description: "The category IDs excluded from the report",
			},
			"questionIDs": &graphql.Field{
				Type:        graphql.NewList(graphql.String),
				Description: "The question IDs excluded from the report",
			},
		},
	})

	jocAggregate := func(typeName string) *graphql.Object {
		lcTypeName := strings.ToLower(typeName)
		return graphql.NewObject(graphql.ObjectConfig{
			Name:        fmt.Sprintf("%sBeneficiaryAgg", typeName),
			Description: fmt.Sprintf("Aggregates a %s over multiple beneficiaries", lcTypeName),
			Fields: graphql.Fields{
				fmt.Sprintf("%sID", lcTypeName): &graphql.Field{
					Type:        graphql.NewNonNull(graphql.String),
					Description: fmt.Sprintf("The ID of the %s being aggregated", lcTypeName),
				},
				"value": &graphql.Field{
					Type:        graphql.Float,
					Description: "The aggregated value",
				},
				"beneficiaryIDs": &graphql.Field{
					Type:        graphql.NewNonNull(graphql.NewList(graphql.String)),
					Description: "The beneficiary IDs included in the aggregation",
				},
				"warnings": &graphql.Field{
					Type:        graphql.NewList(graphql.String),
					Description: "Any warning messages associated with this aggregation. Includes why beneficiaries could not be included",
				},
			},
		})
	}

	jocAggregates := func(typeName string, t graphql.Output) *graphql.Object {
		return graphql.NewObject(graphql.ObjectConfig{
			Name:        fmt.Sprintf("JOC%sAggregations", typeName),
			Description: "Provides aggregations for first and last meetings and the difference between them",
			Fields: graphql.Fields{
				"first": &graphql.Field{
					Description: "Aggregates of the beneficiaries' first meetings",
					Type:        graphql.NewList(t),
				},
				"last": &graphql.Field{
					Description: "Aggregates of the beneficiaries' last meetings in the provided time range",
					Type:        graphql.NewList(t),
				},
				"delta": &graphql.Field{
					Description: "The difference between first and last meetings",
					Type:        graphql.NewList(t),
				},
			},
		})
	}

	return reportTypes{
		JOCType: graphql.NewObject(graphql.ObjectConfig{
			Name:        "JOCServiceReport",
			Description: "This report details journey of change results aggregated across multiple beneficiaries.",
			Fields: graphql.Fields{
				"beneficiaryIDs": &graphql.Field{
					Type:        graphql.NewNonNull(graphql.NewList(graphql.String)),
					Description: "The beneficiary IDs included in the report",
				},
				"questionAggregates": &graphql.Field{
					Type:        graphql.NewNonNull(jocAggregates("Question", jocAggregate("Question"))),
					Description: "Questions aggregated over multiple beneficiaries",
				},
				"categoryAggregates": &graphql.Field{
					Type:        graphql.NewNonNull(jocAggregates("Category", jocAggregate("Category"))),
					Description: "Questions aggregated over multiple beneficiaries",
				},
				"excluded": &graphql.Field{
					Type:        excluded,
					Description: "Any questions or categories excluded from the report. This occurs when there are no instances of them associated with the considered beneficiaries",
				},
				"warnings": &graphql.Field{
					Type:        graphql.NewList(graphql.String),
					Description: "Any warning messages associated with the report. Contains users which could not be included.",
				},
			},
		}),
	}
}

func (v *v1) getRepQueries(repTypes reportTypes) graphql.Fields {
	return graphql.Fields{
		"JOCServiceReport": &graphql.Field{
			Type: repTypes.JOCType,
			Description: `Produces a journey of change report for the organisation between two dates.
This will aggregate questions and categories across multiple beneficiaries.
Beneficiaries with meetings (belonging to the provided question set) within the provided date range will be included in the report.
For each beneficiary, their first meeting (does not have to be in the provided date range) and their last meeting within the provided date range are compared.
Aggregates are calculated over all beneficiaries for the first and last meetings, as well as the difference between them.
`,
			Args: graphql.FieldConfigArgument{
				"start": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "The start of the period to consider when searching for beneficiaries to include in the report. Should be ISO standard timestamp",
				},
				"end": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "The end of the period to consider when searching for beneficiaries to include in the report. Should be ISO standard timestamp",
				},
				"questionSetID": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "The question set to produce the report for",
				},
			},
			Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
				start := p.Args["start"].(string)
				startParsed, err := time.Parse(time.RFC3339, start)
				if err != nil {
					return nil, err
				}
				end := p.Args["end"].(string)
				endParsed, err := time.Parse(time.RFC3339, end)
				if err != nil {
					return nil, err
				}
				osID := p.Args["questionSetID"].(string)
				return logic.GetJOCServiceReport(startParsed, endParsed, osID, v.db, u)
			}),
		},
	}
}
