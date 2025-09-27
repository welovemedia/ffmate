package umami

import (
	"goyave.dev/goyave/v5"
	v "goyave.dev/goyave/v5/validation"
)

func (c *Controller) UmamiRequest(_ *goyave.Request) v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Object()}},
		{Path: "payload.hostname", Rules: v.List{v.String(), v.Required()}},
		{Path: "payload.language", Rules: v.List{v.String(), v.Required()}},
		{Path: "payload.referrer", Rules: v.List{v.String(), v.Required()}},
		{Path: "payload.screen", Rules: v.List{v.String(), v.Required()}},
		{Path: "payload.title", Rules: v.List{v.String(), v.Required()}},
		{Path: "payload.url", Rules: v.List{v.String(), v.Required()}},
		{Path: "payload.website", Rules: v.List{v.String(), v.Required()}},
	}
}
