package api

import (
	"time"

	"github.com/graphql-go/graphql"
	impact "github.com/impactasaurus/server"
	"github.com/impactasaurus/server/auth"
)

func (v *v1) getSchema(orgTypes organisationTypes, osTypes outcomeSetTypes, meetTypes meetingTypes) (*graphql.Schema, error) {

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"outcomesets": &graphql.Field{
				Type:        graphql.NewList(osTypes.outcomeSetType),
				Description: "Gather all outcome sets",
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					return v.db.GetOutcomeSets(u)
				}),
			},
			"outcomeset": &graphql.Field{
				Type:        osTypes.outcomeSetType,
				Description: "Gather a specific outcome set",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Description: "The ID of the outcomeset",
						Type:        graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					return v.db.GetOutcomeSet(p.Args["id"].(string), u)
				}),
			},
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
			"meeting": &graphql.Field{
				Type:        meetTypes.meetingType,
				Description: "Get a meeting by meeting ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Description: "The ID of the meeting",
						Type:        graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					return v.db.GetMeeting(p.Args["id"].(string), u)
				}),
			},
			"meetings": &graphql.Field{
				Type:        graphql.NewList(meetTypes.meetingType),
				Description: "Get all meetings associated with a beneficiary",
				Args: graphql.FieldConfigArgument{
					"beneficiary": &graphql.ArgumentConfig{
						Description: "The ID of the beneficiary",
						Type:        graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					return v.db.GetMeetingsForBeneficiary(p.Args["beneficiary"].(string), u)
				}),
			},
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"AddOutcomeSet": &graphql.Field{
				Type:        osTypes.outcomeSetType,
				Description: "Create a new outcomeset",
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The name of the outcomeset",
					},
					"description": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "An optional description",
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					name := p.Args["name"].(string)
					description := getNullableString(p.Args, "description")
					return v.db.NewOutcomeSet(name, description, u)
				}),
			},
			"EditOutcomeSet": &graphql.Field{
				Type:        osTypes.outcomeSetType,
				Description: "Edit an outcomeset",
				Args: graphql.FieldConfigArgument{
					"outcomeSetID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.ID),
						Description: "The ID of the outcomeset",
					},
					"name": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The new name to apply to the outcomeset",
					},
					"description": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "The new description to apply to the outcomeset, if left null, any existing description will be removed",
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					id := p.Args["outcomeSetID"].(string)
					name := p.Args["name"].(string)
					description := getNullableString(p.Args, "description")
					return v.db.EditOutcomeSet(id, name, description, u)
				}),
			},
			"DeleteOutcomeSet": &graphql.Field{
				Type:        graphql.ID,
				Description: "Deletes an outcomeset and returns the ID of the deleted outcomeset",
				Args: graphql.FieldConfigArgument{
					"outcomeSetID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.ID),
						Description: "The ID of the outcomeset",
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					id := p.Args["outcomeSetID"].(string)
					if err := v.db.DeleteOutcomeSet(id, u); err != nil {
						return nil, err
					}
					return id, nil
				}),
			},
			"AddCategory": &graphql.Field{
				Type:        osTypes.outcomeSetType,
				Description: "Add a category to the outcome set",
				Args: graphql.FieldConfigArgument{
					"outcomeSetID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.ID),
						Description: "The ID of the outcomeset",
					},
					"name": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "Name of the category",
					},
					"description": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Description of the category",
					},
					"aggregation": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(osTypes.aggregationEnum),
						Description: "The aggregation applied to the category",
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					id := p.Args["outcomeSetID"].(string)
					name := p.Args["name"].(string)
					description := getNullableString(p.Args, "description")
					aggregation := p.Args["aggregation"].(impact.Aggregation)
					if _, err := v.db.NewCategory(id, name, description, aggregation, u); err != nil {
						return nil, err
					}
					return v.db.GetOutcomeSet(id, u)
				}),
			},
			"DeleteCategory": &graphql.Field{
				Type:        osTypes.outcomeSetType,
				Description: "Remove a category from an outcome set. The category being removed must not be applied to any questions.",
				Args: graphql.FieldConfigArgument{
					"outcomeSetID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The ID of the outcomeset",
					},
					"categoryID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The ID of the category",
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					outcomeSetID := p.Args["outcomeSetID"].(string)
					categoryID := p.Args["categoryID"].(string)
					if err := v.db.DeleteCategory(outcomeSetID, categoryID, u); err != nil {
						return nil, err
					}
					return v.db.GetOutcomeSet(outcomeSetID, u)
				}),
			},
			"SetCategory": &graphql.Field{
				Type:        osTypes.outcomeSetType,
				Description: "Set or remove the category associated with a question.",
				Args: graphql.FieldConfigArgument{
					"outcomeSetID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The ID of the outcomeset",
					},
					"questionID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The ID of the question",
					},
					"categoryID": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "The ID of the category. If NULL, the category associated with the question is removed",
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					outcomeSetID := p.Args["outcomeSetID"].(string)
					questionID := p.Args["questionID"].(string)
					categoryID := getNullableString(p.Args, "categoryID")
					var dbErr error
					if categoryID == "" {
						_, dbErr = v.db.RemoveCategory(outcomeSetID, questionID, u)
					} else {
						_, dbErr = v.db.SetCategory(outcomeSetID, questionID, categoryID, u)
					}
					if dbErr != nil {
						return nil, dbErr
					}
					return v.db.GetOutcomeSet(outcomeSetID, u)
				}),
			},
			"AddLikertQuestion": &graphql.Field{
				Type:        osTypes.outcomeSetType,
				Description: "Add a likert scale question to an outcome set",
				Args: graphql.FieldConfigArgument{
					"outcomeSetID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.ID),
						Description: "The ID of the outcomeset",
					},
					"question": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "Question to be asked",
					},
					"minValue": &graphql.ArgumentConfig{
						Type:        graphql.Int,
						Description: "Minimum value of the likert scale",
					},
					"maxValue": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.Int),
						Description: "Maximum value of the likert scale",
					},
					"minLabel": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Label associated with the minimum value of the likert scale",
					},
					"maxLabel": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Label associated with the maximum value of the likert scale",
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					id := p.Args["outcomeSetID"].(string)
					question := p.Args["question"].(string)
					minValue := getNullableInt(p.Args, "minValue")
					maxValue := p.Args["maxValue"].(int)
					minLabel := getNullableString(p.Args, "minLabel")
					maxLabel := getNullableString(p.Args, "maxLabel")
					if _, err := v.db.NewQuestion(id, question, impact.LIKERT, map[string]interface{}{
						"minValue": minValue,
						"maxValue": maxValue,
						"minLabel": minLabel,
						"maxLabel": maxLabel,
					}, u); err != nil {
						return nil, err
					}
					return v.db.GetOutcomeSet(id, u)
				}),
			},
			"DeleteQuestion": &graphql.Field{
				Type:        osTypes.outcomeSetType,
				Description: "Remove a question from an outcome set",
				Args: graphql.FieldConfigArgument{
					"outcomeSetID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The ID of the outcomeset",
					},
					"questionID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The ID of the question",
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					outcomeSetID := p.Args["outcomeSetID"].(string)
					questionID := p.Args["questionID"].(string)
					if err := v.db.DeleteQuestion(outcomeSetID, questionID, u); err != nil {
						return nil, err
					}
					return v.db.GetOutcomeSet(outcomeSetID, u)
				}),
			},
			"AddMeeting": &graphql.Field{
				Type:        meetTypes.meetingType,
				Description: "Create a new meeting",
				Args: graphql.FieldConfigArgument{
					"beneficiaryID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The ID associated with the beneficiary being interviewed",
					},
					"outcomeSetID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The ID of the outcome set being used",
					},
					"conducted": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The time and date when the meeting was conducted. Should be ISO standard timestamp",
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					beneficiaryID := p.Args["beneficiaryID"].(string)
					outcomeSetID := p.Args["outcomeSetID"].(string)
					conducted := p.Args["conducted"].(string)
					parsedConducted, err := time.Parse(time.RFC3339, conducted)
					if err != nil {
						return nil, err
					}
					return v.db.NewMeeting(beneficiaryID, outcomeSetID, parsedConducted, u)
				}),
			},
			"AddLikertAnswer": &graphql.Field{
				Type:        meetTypes.meetingType,
				Description: "Provide an answer for a Likert Scale question",
				Args: graphql.FieldConfigArgument{
					"meetingID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The ID of the meeting the answer is associated with",
					},
					"questionID": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "The ID of the question being answered",
					},
					"value": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.Int),
						Description: "The value given for the particular likert scale",
					},
				},
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					meetingID := p.Args["meetingID"].(string)
					questionID := p.Args["questionID"].(string)
					value := p.Args["value"].(int)
					return v.db.NewAnswer(meetingID, impact.Answer{
						QuestionID: questionID,
						Type:       impact.INT,
						Answer:     value,
					}, u)
				}),
			},
			//"DeleteMeeting",
		},
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
