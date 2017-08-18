package api

import (
	"errors"
	"github.com/graphql-go/graphql"
	impact "github.com/impactasaurus/server"
	"github.com/impactasaurus/server/auth"
)

func (v *v1) initOutcomeSetTypes(orgTypes organisationTypes) outcomeSetTypes {
	ret := outcomeSetTypes{}

	ret.questionInterface = graphql.NewInterface(graphql.InterfaceConfig{
		Name:        "QuestionInterface",
		Description: "The interface satisfied by all question types",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.ID),
				Description: "Unique ID for the question",
			},
			"question": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The question",
			},
			"archived": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Whether the question has been archived",
			},
			"categoryID": &graphql.Field{
				Type:        graphql.String,
				Description: "The category the question belongs to",
			},
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			obj, ok := p.Value.(impact.Question)
			if !ok {
				return ret.likertScale
			}
			switch obj.Type {
			case impact.LIKERT:
				return ret.likertScale
			default:
				return ret.likertScale
			}
		},
	})

	ret.likertScale = graphql.NewObject(graphql.ObjectConfig{
		Name:        "LikertScale",
		Description: "Question gathering information using Likert Scales",
		Interfaces: []*graphql.Interface{
			ret.questionInterface,
		},
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.ID),
				Description: "Unique ID for the question",
			},
			"question": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The question",
			},
			"archived": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Whether the question has been archived",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					obj, ok := p.Source.(impact.Question)
					if !ok {
						return nil, errors.New("Expecting an impact.Question")
					}
					return obj.Deleted, nil
				},
			},
			"categoryID": &graphql.Field{
				Type:        graphql.String,
				Description: "The category the question belongs to",
			},
			"minValue": &graphql.Field{
				Type:        graphql.Int,
				Description: "The minimum value in the scale",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					obj, ok := p.Source.(impact.Question)
					if !ok {
						return nil, errors.New("Expecting an impact.Question")
					}
					minValue, ok := obj.Options["minValue"]
					if !ok {
						return nil, nil
					}
					minValueInt, ok := minValue.(int)
					if !ok {
						return nil, errors.New("Min likert value should be an int")
					}
					return minValueInt, nil
				},
			},
			"maxValue": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "The maximum value in the scale",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					obj, ok := p.Source.(impact.Question)
					if !ok {
						return nil, errors.New("Expecting an impact.Question")
					}
					maxValue, ok := obj.Options["maxValue"]
					if !ok {
						return nil, nil
					}
					maxValueInt, ok := maxValue.(int)
					if !ok {
						return nil, errors.New("Max likert value should be an int")
					}
					return maxValueInt, nil
				},
			},
			"minLabel": &graphql.Field{
				Type:        graphql.String,
				Description: "The string labelling the minimum value in the scale",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					obj, ok := p.Source.(impact.Question)
					if !ok {
						return nil, errors.New("Expecting an impact.Question")
					}
					label, ok := obj.Options["minLabel"]
					if !ok {
						return nil, nil
					}
					labelStr, ok := label.(string)
					if !ok {
						return nil, errors.New("Min likert label should be an string")
					}
					return labelStr, nil
				},
			},
			"maxLabel": &graphql.Field{
				Type:        graphql.String,
				Description: "The string labelling the maximum value in the scale",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					obj, ok := p.Source.(impact.Question)
					if !ok {
						return nil, errors.New("Expecting an impact.Question")
					}
					label, ok := obj.Options["maxLabel"]
					if !ok {
						return nil, nil
					}
					labelStr, ok := label.(string)
					if !ok {
						return nil, errors.New("Max likert label should be an string")
					}
					return labelStr, nil
				},
			},
		},
	})

	ret.aggregationEnum = graphql.NewEnum(graphql.EnumConfig{
		Name:        "Aggregation",
		Description: "Aggregation functions available",
		Values: graphql.EnumValueConfigMap{
			string(impact.MEAN): &graphql.EnumValueConfig{
				Value:       impact.MEAN,
				Description: "Mean",
			},
			string(impact.SUM): &graphql.EnumValueConfig{
				Value:       impact.SUM,
				Description: "Sum",
			},
		},
	})

	ret.categoryType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Category",
		Description: "Categorises a set of questions. Used for aggregation",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.ID),
				Description: "Unique ID",
			},
			"name": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Name of the category",
			},
			"description": &graphql.Field{
				Type:        graphql.String,
				Description: "Description of the category",
			},
			"aggregation": &graphql.Field{
				Type:        graphql.NewNonNull(ret.aggregationEnum),
				Description: "The aggregation applied to the category",
			},
		},
	})

	ret.outcomeSetType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "OutcomeSet",
		Description: "A set of questions to determine outcomes",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.ID),
				Description: "Unique ID",
			},
			"organisationID": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Organisation's unique ID",
			},
			"organisation": &graphql.Field{
				Type:        graphql.NewNonNull(orgTypes.organisationType),
				Description: "The owning organisation of the outcome set",
				Resolve: userRestrictedResolver(func(p graphql.ResolveParams, u auth.User) (interface{}, error) {
					obj, ok := p.Source.(impact.OutcomeSet)
					if !ok {
						return nil, errors.New("Expecting an impact.Meeting")
					}
					return v.db.GetOrganisation(obj.OrganisationID, u)
				}),
			},
			"name": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Name of the outcome set",
			},
			"description": &graphql.Field{
				Type:        graphql.String,
				Description: "Information about the outcome set",
			},
			"questions": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(ret.questionInterface)),
				Description: "Questions associated with the outcome set",
			},
			"categories": &graphql.Field{
				Type:        graphql.NewList(ret.categoryType),
				Description: "Questions associated with the outcome set",
			},
		},
	})

	return ret
}
