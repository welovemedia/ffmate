package preset

import (
	"github.com/welovemedia/ffmate/v2/internal/validate"
	"goyave.dev/goyave/v5"
	v "goyave.dev/goyave/v5/validation"
)

func (ctrl *Controller) NewPresetRequest(_ *goyave.Request) v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Object()}},
		{Path: "name", Rules: v.List{v.String(), v.Required()}},
		{Path: "description", Rules: v.List{v.String()}},
		{Path: "command", Rules: v.List{v.String(), v.Required()}},
		{Path: "priority", Rules: v.List{v.Uint()}},
		{Path: "outputFile", Rules: v.List{v.String()}},
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
		{Path: "globalPresetName", Rules: v.List{v.String()}},
	}
}
