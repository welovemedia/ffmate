package task

import (
	"github.com/welovemedia/ffmate/v2/internal/validate"
	"goyave.dev/goyave/v5"
	v "goyave.dev/goyave/v5/validation"
)

func (ctrl *Controller) NewTaskRequest(_ *goyave.Request) v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Object()}},
		{Path: "name", Rules: v.List{v.String(), v.Required()}},

		{Path: "command", Rules: v.List{
			v.String(),
			v.WithMessage(v.RequiredIf(func(ctx *v.Context) bool {
				data, ok := ctx.Data.(map[string]interface{})
				if !ok {
					return false
				}
				if val, exists := data["preset"]; !exists || val == nil || val == "" {
					return true
				}
				return false
			}), "Either command or preset must be set"),
		}},

		{Path: "preset", Rules: v.List{
			v.String(),
			v.WithMessage(v.RequiredIf(func(ctx *v.Context) bool {
				data, ok := ctx.Data.(map[string]interface{})
				if !ok {
					return false
				}
				if val, exists := data["command"]; !exists || val == nil || val == "" {
					return true
				}
				return false
			}), "Either command or preset must be set"),
		}},

		{Path: "priority", Rules: v.List{v.Uint()}},

		{Path: "inputFile", Rules: v.List{v.String()}},
		{Path: "outputFile", Rules: v.List{v.String()}},

		{Path: "metadata", Rules: v.List{v.Object()}},

		{Path: "webhooks", Rules: v.List{v.Array()}},
		{Path: "webhooks[].url", Rules: v.List{validate.PreserveValue(v.URL()), v.Required()}},
		{Path: "webhooks[].event", Rules: v.List{v.String(), v.Required()}},

		{Path: "preProcessing", Rules: v.List{v.Object()}},
		{Path: "preProcessing.scriptPath", Rules: v.List{v.String()}},
		{Path: "preProcessing.sidecarPath", Rules: v.List{v.String()}},
		{Path: "preProcessing.importSidecar", Rules: v.List{v.Bool()}},

		{Path: "postProcessing", Rules: v.List{v.Object()}},
		{Path: "postProcessing.scriptPath", Rules: v.List{v.String()}},
		{Path: "postProcessing.sidecarPath", Rules: v.List{v.String()}},
	}
}
