package models

type GetPaginatedQuery struct {
	Offset int
	Limit  int
	Q      string
}
