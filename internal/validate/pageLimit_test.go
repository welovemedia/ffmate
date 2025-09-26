package validate

import (
	"testing"

	"github.com/stretchr/testify/require"
	v "goyave.dev/goyave/v5/validation"
)

func TestPaginationRequest_Valid(t *testing.T) {
	rules := PaginationRequest(nil)

	input := map[string]any{
		"page":    2,
		"perPage": 50,
	}

	errs, internal := v.Validate(&v.Options{
		Data:  input,
		Rules: rules, // RuleSet implements Ruler (AsRules)
	})
	require.Empty(t, internal, "internal validation errors should be empty")
	require.Nil(t, errs, "expected no validation errors")
}

func TestPaginationRequest_Invalid(t *testing.T) {
	rules := PaginationRequest(nil)

	input := map[string]any{
		"page":    -1,  // invalid: < 0
		"perPage": 200, // invalid: > 100
	}

	errs, internal := v.Validate(&v.Options{
		Data:  input,
		Rules: rules,
	})
	require.Empty(t, internal, "internal validation errors should be empty")
	require.NotNil(t, errs, "expected validation errors")
	// FieldsErrors is map[string]*validation.Errors
	require.NotNil(t, errs.Fields["page"])
	require.NotNil(t, errs.Fields["perPage"])
	require.NotEmpty(t, errs.Fields["page"].Errors)
	require.NotEmpty(t, errs.Fields["perPage"].Errors)
}
