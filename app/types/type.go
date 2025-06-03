package types

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RepoTimestampConfig struct {
	CreatedAt bool // Base repo createdAt timestamp config, default is false
	UpdatedAt bool // Base repo updatedAt timestamp config, default is false
}

type AppDB struct {
	MongoDB *mongo.Database
}

type QueryParamsFilterOp string

const (
	OpEq    QueryParamsFilterOp = "eq"
	OpLike  QueryParamsFilterOp = "like"
	OpGte   QueryParamsFilterOp = "gte"
	OpLte   QueryParamsFilterOp = "lte"
	OpStart QueryParamsFilterOp = "start"
	OpEnd   QueryParamsFilterOp = "end"
)

type QueryParamsFilter struct {
	Field    string
	Operator QueryParamsFilterOp
	Value    string
}

type QueryParamsSortField struct {
	Field      string
	Descending bool
}

type PaginationParams struct {
	Page    int
	PerPage int
}

type GetAllFiltersAndSorts struct {
	QueryParamsFilters    []QueryParamsFilter
	QueryParamsSortFields []QueryParamsSortField
}

type PaginatedRecords[T any] struct {
	Records    []T `json:"records"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
