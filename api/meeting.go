package api

import (
	"errors"
	"github.com/graphql-go/graphql"
	impact "github.com/impactasaurus/server"
	"github.com/impactasaurus/server/auth"
	"github.com/impactasaurus/server/logic"
	"time"
)

func (v *v1) initMeetingTypes(orgTypes organisationTypes, osTypes outcomeSetTypes) meetingTypes {
	ret := meetingTypes{}

	ret.answerInterface = graphql.NewInterface(graphql.InterfaceConfig{
		Name:        "AnswerInterface",
		Description: "The interface satisfied by all answer types",
		Fields: graphql.Fields{
			"questionID": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The ID of the question answered",
			},
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			obj, ok := p.Value.(impact.Answer)
			if !ok {
				return ret.intAnswer
			}
			switch obj.Type {
			case impact.INT:
				return ret.intAnswer
			default:
				return ret.intAnswer
			}
		},
	})

	ret.intAnswer = graphql.NewObject(graphql.ObjectConfig{
		Name:        "IntAnswer",
		Description: "Answer containing an integer value",
		Interfaces: []*graphql.Interface{
			ret.answerInterface,
		},
		Fields: graphql.Fields{
			"questionID": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The ID of the question answered",
			},
			"answer": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "The provided int answer",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					obj, ok := p.Source.(impact.Answer)
					if !ok {
						return nil, errors.New("Expecting an impact.Answer")
					}
					num, ok := obj.Answer.(int)
					if !ok {
						return nil, errors.New("Expected an int value")
					}
					return num, nil
				},
			},
		},
	})

	ret.categoryAggregate = graphql.NewObject(graphql.ObjectConfig{
		Name:        "CategoryAggregate",
		Description: "An aggregation of answers to the category level",
		Fields: graphql.Fields{
			"categoryID": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The ID of the category being aggregated",
			},
			"value": &graphql.Field{
				Type:        graphql.Float,
				Description: "The aggregated value",
			},
		},
	})

	ret.aggregates = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Aggregates",
		Description: "Aggregations of the meeting",
		Fields: graphql.Fields{
			"category": &graphql.Field{
				Type:        graphql.NewList(ret.categoryAggregate),
				Description: "Answers aggregated to the category level",
			},
		},
	})

	ret.meetingType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Meeting",
		Description: "A set of answers for an outcome set",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.ID),
				Description: "Unique ID for the meeting",
			},
			"beneficiary": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The beneficiary providing the answers",
			},
			"user": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The user who collected the answers",
			},
			"outcomeSetID": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The ID of the outcome set answered",
			},
			"outcomeSet": &graphql.Field{
				Type:        graphql.NewNonNull(osTypes.outcomeSetType),
				Description: "The outcome set answered",
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					obj, ok := p.Source.(impact.Meeting)
					if !ok {
						return nil, errors.New("Expecting an impact.Meeting")
					}
					return v.db.GetOutcomeSet(obj.OutcomeSetID, u)
				}),
			},
			"organisationID": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Organisation's unique ID",
			},
			"organisation": &graphql.Field{
				Type:        graphql.NewNonNull(orgTypes.organisationType),
				Description: "The owning organisation of the outcome set",
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					obj, ok := p.Source.(impact.Meeting)
					if !ok {
						return nil, errors.New("Expecting an impact.Meeting")
					}
					return v.db.GetOrganisation(obj.OrganisationID, u)
				}),
			},
			"answers": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(ret.answerInterface)),
				Description: "The answers provided in the meeting",
			},
			"aggregates": &graphql.Field{
				Type:        ret.aggregates,
				Description: "Aggregations of the meeting's answers",
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					obj, ok := p.Source.(impact.Meeting)
					if !ok {
						return nil, errors.New("Expecting an impact.Meeting")
					}
					os, err := v.db.GetOutcomeSet(obj.OutcomeSetID, u)
					if err != nil {
						return nil, err
					}
					catAgs, err := logic.GetCategoryAggregates(obj, os)
					if err != nil {
						return nil, err
					}
					return impact.Aggregates{
						Category: catAgs,
					}, nil
				}),
			},
			"conducted": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "When the meeting was conducted",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					obj, ok := p.Source.(impact.Meeting)
					if !ok {
						return nil, errors.New("Expecting an impact.Meeting")
					}
					return obj.Conducted.Format(time.RFC3339), nil
				},
			},
			"created": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "When the meeting was entered into the system",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					obj, ok := p.Source.(impact.Meeting)
					if !ok {
						return nil, errors.New("Expecting an impact.Meeting")
					}
					return obj.Created.Format(time.RFC3339), nil
				},
			},
			"modified": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "When the meeting was last modified in the system",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					obj, ok := p.Source.(impact.Meeting)
					if !ok {
						return nil, errors.New("Expecting an impact.Meeting")
					}
					return obj.Modified.Format(time.RFC3339), nil
				},
			},
		},
	})
	return ret
}

func (v *v1) getMeetingQueries(meetTypes meetingTypes) graphql.Fields {
	return graphql.Fields{
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
	}
}

func (v *v1) getMeetingMutations(meetTypes meetingTypes) graphql.Fields {
	return graphql.Fields{
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
	}
}
