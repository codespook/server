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

func (v *v1) getOSQueries(osTypes outcomeSetTypes) graphql.Fields {
	return graphql.Fields{
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
	}
}

func (v *v1) getOSMutations(osTypes outcomeSetTypes) graphql.Fields {
	return graphql.Fields{
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
	}
}
