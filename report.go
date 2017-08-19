package server

// BenAgg stores aggregations of values over multiple beneficiaries
type BenAgg struct {
	Value          float32  `json:"value"`
	BeneficiaryIDs []string `json:"beneficiaryIDs"`
	Warnings       []string `json:"warnings"`
}

// CatBenAgg is a BenAgg associated with a question category
type CatBenAgg struct {
	CategoryID string `json:"categoryID"`
	BenAgg
}

// QBenAgg is a BenAgg associated with a question
type QBenAgg struct {
	QuestionID string `json:"questionID"`
	BenAgg
}

type Excluded struct {
	CategoryIDs []string `json:"categoryIDs"`
	QuestionIDs []string `json:"questionIDs"`
}

type JOCCatAggs struct {
	First []CatBenAgg `json:"first"`
	Last  []CatBenAgg `json:"last"`
	Delta []CatBenAgg `json:"delta"`
}

type JOCQAggs struct {
	First []QBenAgg `json:"first"`
	Last  []QBenAgg `json:"last"`
	Delta []QBenAgg `json:"delta"`
}

type JOCServiceReport struct {
	BeneficiaryIDs     []string   `json:"beneficiaryIDs"`
	QuestionAggregates JOCQAggs   `json:"questionAggregates"`
	CategoryAggregates JOCCatAggs `json:"categoryAggregates"`
	Excluded           Excluded   `json:"excluded"`
}
