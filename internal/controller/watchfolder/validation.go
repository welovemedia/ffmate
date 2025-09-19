package watchfolder

import (
	"goyave.dev/goyave/v5"
	v "goyave.dev/goyave/v5/validation"
)

func (ctrl *Controller) NewWatchfolderRequest(_ *goyave.Request) v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Object()}},
		{Path: "name", Rules: v.List{v.String(), v.Required()}},
		{Path: "description", Rules: v.List{v.String()}},
		{Path: "path", Rules: v.List{v.String(), v.Required()}},
		{Path: "interval", Rules: v.List{v.Int(), v.Required()}},
		{Path: "growthChecks", Rules: v.List{v.Int(), v.Required()}},
		{Path: "suspended", Rules: v.List{v.Bool()}},
		{Path: "preset", Rules: v.List{v.String(), v.Required()}},
		{Path: "filter", Rules: v.List{v.Object()}},
		{Path: "filter.extensions", Rules: v.List{v.Object()}},
		{Path: "filter.extensions.exclude[]", Rules: v.List{v.String()}},
		{Path: "filter.extensions.include[]", Rules: v.List{v.String()}},
	}
}
