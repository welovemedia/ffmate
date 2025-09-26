package validate

import (
	"testing"

	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5/validation"
)

type mockValidator struct {
	validation.BaseValidator
	mutateValue   func(any) any
	placeholders  []string
	name          string
	typeDependent bool
	typeValidator bool
	called        bool
	returns       bool
}

func (m *mockValidator) Init(opts *validation.Options) {
	m.BaseValidator.Init(opts)
}

func (m *mockValidator) Name() string {
	return m.name
}

func (m *mockValidator) IsTypeDependent() bool {
	return m.typeDependent
}

func (m *mockValidator) IsType() bool {
	return m.typeValidator
}

func (m *mockValidator) MessagePlaceholders(_ *validation.Context) []string {
	return m.placeholders
}

func (m *mockValidator) Validate(ctx *validation.Context) bool {
	m.called = true
	if m.mutateValue != nil {
		ctx.Value = m.mutateValue(ctx.Value)
	}
	return m.returns
}

func TestPreserveValueValidator_WithInner(t *testing.T) {
	inner := &mockValidator{
		name:          "mock",
		typeDependent: true,
		typeValidator: true,
		returns:       true,
		placeholders:  []string{"foo", "bar"},
		mutateValue: func(_ any) any {
			// simulate inner validator changing the value
			return "mutated"
		},
	}

	v := PreserveValue(inner)
	require.Equal(t, "mock", v.Name())
	require.True(t, v.IsTypeDependent())
	require.True(t, v.IsType())

	ctx := &validation.Context{Value: "original"}
	result := v.Validate(ctx)

	require.True(t, inner.called, "inner validator should be called")
	require.True(t, result, "result should come from inner validator")
	require.Equal(t, "original", ctx.Value, "ctx.Value must be restored after validation")
	require.Equal(t, []string{"foo", "bar"}, v.MessagePlaceholders(ctx))
}

func TestPreserveValueValidator_NoInner(t *testing.T) {
	v := PreserveValue(nil)
	ctx := &validation.Context{Value: "keep"}
	require.Equal(t, "preserve_value", v.Name())
	require.False(t, v.IsTypeDependent())
	require.False(t, v.IsType())
	require.Nil(t, v.MessagePlaceholders(ctx))
	result := v.Validate(ctx)
	require.True(t, result, "validator with nil inner must always succeed")
	require.Equal(t, "keep", ctx.Value, "ctx.Value should be untouched")
}
