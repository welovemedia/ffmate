package testsuite

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockReadCloser struct {
	err  error
	data []byte
}

func (m *mockReadCloser) Read(p []byte) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return bytes.NewReader(m.data).Read(p)
}

func (m *mockReadCloser) Close() error { return nil }

func TestParseJSONBody(t *testing.T) {
	type sample struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name      string
		body      io.ReadCloser
		expect    sample
		expectErr bool
	}{
		{
			name:   "valid JSON",
			body:   io.NopCloser(bytes.NewBufferString(`{"name":"Alice","age":30}`)),
			expect: sample{Name: "Alice", Age: 30},
		},
		{
			name:      "invalid JSON",
			body:      io.NopCloser(bytes.NewBufferString(`{"name":Alice,"age":30}`)),
			expectErr: true,
		},
		{
			name:      "read error",
			body:      &mockReadCloser{err: errors.New("read error")},
			expectErr: true,
		},
		{
			name:   "empty body (zero value)",
			body:   io.NopCloser(bytes.NewBuffer(nil)),
			expect: sample{}, // Should convert to zero value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseJSONBody[sample](tt.body)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect, result)
			}
		})
	}
}

func TestParseBody(t *testing.T) {
	tests := []struct {
		name      string
		body      io.ReadCloser
		expect    []byte
		expectErr bool
	}{
		{
			name:   "valid body",
			body:   io.NopCloser(bytes.NewBufferString("hello world")),
			expect: []byte("hello world"),
		},
		{
			name:      "read error",
			body:      &mockReadCloser{err: errors.New("read error")},
			expectErr: true,
		},
		{
			name:   "empty body",
			body:   io.NopCloser(bytes.NewBuffer(nil)),
			expect: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseBody(tt.body)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect, result)
			}
		})
	}
}
