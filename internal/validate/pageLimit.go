package validate

import (
	"goyave.dev/goyave/v5"
	v "goyave.dev/goyave/v5/validation"
)

func PaginationRequest(_ *goyave.Request) v.RuleSet {
	return v.RuleSet{
		{Path: "page", Rules: v.List{v.Int(), v.Min(0)}},
		{Path: "perPage", Rules: v.List{v.Int(), v.Between(1, 100)}},
	}
}
