package webhook

import (
	"github.com/welovemedia/ffmate/internal/validate"
	"goyave.dev/goyave/v5"
	v "goyave.dev/goyave/v5/validation"
)

func (ctrl *Controller) NewWebhookRequest(_ *goyave.Request) v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Object()}},
		{Path: "event", Rules: v.List{v.String(), v.Required()}},
		{Path: "url", Rules: v.List{validate.PreserveValue(v.URL()), v.Required()}},
	}
}
