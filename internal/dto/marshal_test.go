package dto

import (
	"database/sql/driver"
	"reflect"
	"strings"
	"testing"
	"time"
)

type valuer interface {
	Value() (driver.Value, error)
}

type scanner interface {
	Scan(obj any) error
}

func mustValue(t *testing.T, v valuer) []byte {
	t.Helper()
	raw, err := v.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	b, ok := raw.([]byte)
	if !ok {
		t.Fatalf("expected []byte from Value(), got %T", raw)
	}
	return b
}

func derefValue(s scanner) any {
	return reflect.ValueOf(s).Elem().Interface()
}

func setValue(dst scanner, src any) {
	reflect.ValueOf(dst).Elem().Set(reflect.ValueOf(src))
}

func TestMarshalScan_RoundTrip(t *testing.T) {
	// Fixed timestamps to avoid monotonic clock issues in comparisons
	t1 := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	t2 := time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC)

	cases := []struct {
		name     string
		original valuer
		zero     func() scanner
	}{
		{
			name:     "WatchfolderFilter",
			original: WatchfolderFilter{Extensions: &WatchfolderFilterExtensions{Include: []string{"mp4", "mkv"}, Exclude: []string{"tmp"}}},
			zero: func() scanner {
				var dst WatchfolderFilter
				return &dst
			},
		},
		{
			name:     "Settings",
			original: Settings{},
			zero: func() scanner {
				var dst Settings
				return &dst
			},
		},
		{
			name:     "MetadataMap",
			original: MetadataMap{"a": "b", "num": float64(1.23), "flag": true, "nested": map[string]any{"k": "v"}, "arr": []any{"x", "y"}},
			zero: func() scanner {
				var dst MetadataMap
				return &dst
			},
		},
		{
			name:     "WebhookResponse",
			original: WebhookResponse{Headers: map[string][]string{"Content-Type": {"application/json"}}, Body: `{"ok":true}`, Status: 200},
			zero: func() scanner {
				var dst WebhookResponse
				return &dst
			},
		},
		{
			name:     "WebhookRequest",
			original: WebhookRequest{Headers: map[string][]string{"X-Test": {"1", "2"}}, Body: "payload"},
			zero: func() scanner {
				var dst WebhookRequest
				return &dst
			},
		},
		{
			name:     "Webhook",
			original: Webhook{CreatedAt: t1, UpdatedAt: t2, Event: TaskCreated, URL: "https://example.com/hook", UUID: "wh-123"},
			zero: func() scanner {
				var dst Webhook
				return &dst
			},
		},
		{
			name: "WebhookExecution",
			original: WebhookExecution{
				CreatedAt: t1,
				UpdatedAt: t2,
				Request:   &WebhookRequest{Headers: map[string][]string{"A": {"B"}}, Body: "req"},
				Response:  &WebhookResponse{Headers: map[string][]string{"C": {"D"}}, Body: "res", Status: 202},
				UUID:      "exec-1",
				Event:     TaskUpdated,
				URL:       "https://example.com/hook",
			},
			zero: func() scanner {
				var dst WebhookExecution
				return &dst
			},
		},
		{
			name:     "DirectWebhooks",
			original: DirectWebhooks{{Event: TaskCreated, URL: "https://ex.com/a"}, {Event: TaskDeleted, URL: "https://ex.com/b"}},
			zero: func() scanner {
				var dst DirectWebhooks
				return &dst
			},
		},
		{
			name:     "NewWebhook",
			original: NewWebhook{Event: PresetCreated, URL: "https://example.net/new"},
			zero: func() scanner {
				var dst NewWebhook
				return &dst
			},
		},
		{
			name:     "PrePostProcessing",
			original: PrePostProcessing{ScriptPath: &RawResolved{Raw: "echo", Resolved: "/bin/echo"}, SidecarPath: &RawResolved{Raw: "s", Resolved: "/tmp/s"}, Error: "", StartedAt: 10, FinishedAt: 20, ImportSidecar: true},
			zero: func() scanner {
				var dst PrePostProcessing
				return &dst
			},
		},
		{
			name:     "RawResolved",
			original: RawResolved{Raw: "X", Resolved: "Y"},
			zero: func() scanner {
				var dst RawResolved
				return &dst
			},
		},
		{
			name:     "NewPrePostProcessing",
			original: NewPrePostProcessing{ScriptPath: "echo", SidecarPath: "/tmp/s", ImportSidecar: true},
			zero: func() scanner {
				var dst NewPrePostProcessing
				return &dst
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bytes := mustValue(t, c.original)

			dst := c.zero()
			if err := dst.Scan(bytes); err != nil {
				t.Fatalf("Scan() error: %v", err)
			}

			got := derefValue(dst)
			exp := reflect.ValueOf(c.original).Interface()

			if !reflect.DeepEqual(got, exp) {
				t.Fatalf("round-trip mismatch:\nexpected: %#v\ngot:      %#v", exp, got)
			}
		})
	}
}

func TestScan_NilValue_NoChange(t *testing.T) {
	// Use the same cases as round-trip, but we only need zero + set to non-zero and then Scan(nil)
	t1 := time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC)
	t2 := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)

	cases := []struct {
		name     string
		original any
		zero     func() scanner
	}{
		{
			name:     "WatchfolderFilter",
			original: WatchfolderFilter{Extensions: &WatchfolderFilterExtensions{Include: []string{"a"}, Exclude: []string{"b"}}},
			zero: func() scanner {
				var dst WatchfolderFilter
				return &dst
			},
		},
		{
			name:     "Settings",
			original: Settings{},
			zero: func() scanner {
				var dst Settings
				return &dst
			},
		},
		{
			name:     "MetadataMap",
			original: MetadataMap{"k": "v", "n": float64(42)},
			zero: func() scanner {
				var dst MetadataMap
				return &dst
			},
		},
		{
			name:     "WebhookResponse",
			original: WebhookResponse{Headers: map[string][]string{"H": {"1"}}, Body: "b", Status: 201},
			zero: func() scanner {
				var dst WebhookResponse
				return &dst
			},
		},
		{
			name:     "WebhookRequest",
			original: WebhookRequest{Headers: map[string][]string{"Z": {"x", "y"}}, Body: "req"},
			zero: func() scanner {
				var dst WebhookRequest
				return &dst
			},
		},
		{
			name:     "Webhook",
			original: Webhook{CreatedAt: t1, UpdatedAt: t2, Event: WatchfolderUpdated, URL: "u", UUID: "id"},
			zero: func() scanner {
				var dst Webhook
				return &dst
			},
		},
		{
			name: "WebhookExecution",
			original: WebhookExecution{
				CreatedAt: t1,
				UpdatedAt: t2,
				Request:   &WebhookRequest{Headers: map[string][]string{"A": {"B"}}, Body: "r"},
				Response:  &WebhookResponse{Headers: map[string][]string{"C": {"D"}}, Body: "s", Status: 200},
				UUID:      "x",
				Event:     WebhookDeleted,
				URL:       "u",
			},
			zero: func() scanner {
				var dst WebhookExecution
				return &dst
			},
		},
		{
			name:     "DirectWebhooks",
			original: DirectWebhooks{{Event: PresetDeleted, URL: "a"}},
			zero: func() scanner {
				var dst DirectWebhooks
				return &dst
			},
		},
		{
			name:     "NewWebhook",
			original: NewWebhook{Event: BatchFinished, URL: "x"},
			zero: func() scanner {
				var dst NewWebhook
				return &dst
			},
		},
		{
			name:     "PrePostProcessing",
			original: PrePostProcessing{ScriptPath: &RawResolved{Raw: "a", Resolved: "b"}, ImportSidecar: false},
			zero: func() scanner {
				var dst PrePostProcessing
				return &dst
			},
		},
		{
			name:     "RawResolved",
			original: RawResolved{Raw: "r", Resolved: "s"},
			zero: func() scanner {
				var dst RawResolved
				return &dst
			},
		},
		{
			name:     "NewPrePostProcessing",
			original: NewPrePostProcessing{ScriptPath: "p", SidecarPath: "q", ImportSidecar: true},
			zero: func() scanner {
				var dst NewPrePostProcessing
				return &dst
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dst := c.zero()
			setValue(dst, c.original) // make it non-zero

			before := derefValue(dst)
			if err := dst.Scan(nil); err != nil {
				t.Fatalf("Scan(nil) error: %v", err)
			}
			after := derefValue(dst)

			if !reflect.DeepEqual(after, before) {
				t.Fatalf("value changed on Scan(nil):\nbefore: %#v\nafter:  %#v", before, after)
			}
		})
	}
}

func TestScan_InvalidType_Error(t *testing.T) {
	cases := []struct {
		name string
		zero func() scanner
	}{
		{"WatchfolderFilter", func() scanner { var dst WatchfolderFilter; return &dst }},
		{"Settings", func() scanner { var dst Settings; return &dst }},
		{"MetadataMap", func() scanner { var dst MetadataMap; return &dst }},
		{"WebhookResponse", func() scanner { var dst WebhookResponse; return &dst }},
		{"WebhookRequest", func() scanner { var dst WebhookRequest; return &dst }},
		{"Webhook", func() scanner { var dst Webhook; return &dst }},
		{"WebhookExecution", func() scanner { var dst WebhookExecution; return &dst }},
		{"DirectWebhooks", func() scanner { var dst DirectWebhooks; return &dst }},
		{"NewWebhook", func() scanner { var dst NewWebhook; return &dst }},
		{"PrePostProcessing", func() scanner { var dst PrePostProcessing; return &dst }},
		{"RawResolved", func() scanner { var dst RawResolved; return &dst }},
		{"NewPrePostProcessing", func() scanner { var dst NewPrePostProcessing; return &dst }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dst := c.zero()
			err := dst.Scan("not bytes")
			if err == nil {
				t.Fatalf("expected error scanning invalid type, got nil")
			}
			if !strings.Contains(err.Error(), "type assertion to []byte failed") {
				t.Fatalf("unexpected error, want type assertion failure, got: %v", err)
			}
		})
	}
}
