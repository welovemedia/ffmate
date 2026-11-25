package apikey

import (
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"goyave.dev/goyave/v5"
	v "goyave.dev/goyave/v5/validation"
)

func (c *Controller) NewAPIKeyRequest(_ *goyave.Request) v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Object()}},
		{Path: "name", Rules: v.List{v.String(), v.Required(), v.Max(255)}},
		{Path: "access", Rules: v.List{v.String(), v.Required(), v.In([]string{string(dto.APIKeyAccessRead), string(dto.APIKeyAccessWrite)})}},
	}
}
