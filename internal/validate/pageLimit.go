package validate

import (
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"goyave.dev/goyave/v5"
	v "goyave.dev/goyave/v5/validation"
)

func PaginationRequest(_ *goyave.Request) v.RuleSet {
	return v.RuleSet{
		{Path: "page", Rules: v.List{v.Int(), v.Min(0)}},
		{Path: "perPage", Rules: v.List{v.Int(), v.Between(1, 100)}},
	}
}

func PaginationRequestWithTaskFilter(_ *goyave.Request) v.RuleSet {
	return v.RuleSet{
		{Path: "page", Rules: v.List{
			v.Int(),
			v.Min(0),
		}},
		{Path: "perPage", Rules: v.List{
			v.Int(),
			v.Between(1, 100),
		}},
		{Path: "status", Rules: v.List{
			v.String(),
			v.In([]string{
				string(dto.Queued),
				string(dto.Running),
				string(dto.PreProcessing),
				string(dto.PostProcessing),
				string(dto.DoneSuccessful),
				string(dto.DoneError),
				string(dto.DoneCanceled),
				string(dto.All),
			}),
		}},
	}
}
