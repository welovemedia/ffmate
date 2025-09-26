package umami

import (
	"goyave.dev/goyave/v5"
	v "goyave.dev/goyave/v5/validation"
)

func (c *Controller) UmamiRequest(_ *goyave.Request) v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Object()}},
		{Path: "hostname", Rules: v.List{v.String(), v.Required()}},
		{Path: "language", Rules: v.List{v.String(), v.Required()}},
		{Path: "referrer", Rules: v.List{v.String(), v.Required()}},
		{Path: "screen", Rules: v.List{v.String(), v.Required()}},
		{Path: "title", Rules: v.List{v.String(), v.Required()}},
		{Path: "url", Rules: v.List{v.String(), v.Required()}},
		{Path: "website", Rules: v.List{v.String(), v.Required()}},
	}
}
