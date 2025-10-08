package dto

import "goyave.dev/goyave/v5/util/typeutil"

type Pagination struct {
	Page    typeutil.Undefined[int] `json:"page"`
	PerPage typeutil.Undefined[int] `json:"perPage"`
}

type PaginationWithFilter struct {
	Page    typeutil.Undefined[int]        `json:"page"`
	PerPage typeutil.Undefined[int]        `json:"perPage"`
	Status  typeutil.Undefined[TaskStatus] `json:"status"`
}
