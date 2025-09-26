package validate

import (
	"goyave.dev/goyave/v5/validation"
)

type preserveValueValidator struct {
	inner validation.Validator
	validation.BaseValidator
}

func PreserveValue(inner validation.Validator) validation.Validator {
	return &preserveValueValidator{inner: inner}
}

func (v *preserveValueValidator) Init(opts *validation.Options) {
	v.BaseValidator.Init(opts)
	if v.inner != nil {
		v.inner.Init(opts)
	}
}

func (v *preserveValueValidator) Name() string {
	if v.inner == nil {
		return "preserve_value"
	}
	return v.inner.Name()
}

func (v *preserveValueValidator) IsTypeDependent() bool {
	if v.inner == nil {
		return false
	}
	return v.inner.IsTypeDependent()
}

func (v *preserveValueValidator) IsType() bool {
	if v.inner == nil {
		return false
	}
	return v.inner.IsType()
}

func (v *preserveValueValidator) MessagePlaceholders(ctx *validation.Context) []string {
	if v.inner == nil {
		return nil
	}
	return v.inner.MessagePlaceholders(ctx)
}

func (v *preserveValueValidator) Validate(ctx *validation.Context) bool {
	if v.inner == nil {
		return true
	}
	original := ctx.Value
	ok := v.inner.Validate(ctx)
	ctx.Value = original
	return ok
}
