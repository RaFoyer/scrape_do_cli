package outfmt

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestFromFlags(t *testing.T) {
	if _, err := FromFlags(true, true); err == nil {
		t.Fatalf("expected error")
	}
	m, err := FromFlags(true, false)
	if err != nil {
		t.Fatalf("FromFlags: %v", err)
	}
	if !m.JSON || m.Plain {
		t.Fatalf("mode=%+v", m)
	}
}

func TestWriteJSON_Transform(t *testing.T) {
	ctx := context.Background()
	ctx = WithJSONTransform(ctx, JSONTransform{ResultsOnly: true, Select: []string{"id", "nested.value"}})
	payload := map[string]any{
		"results": map[string]any{"id": "x", "nested": map[string]any{"value": 10, "other": true}},
		"meta":    "m",
	}
	var buf bytes.Buffer
	if err := WriteJSON(ctx, &buf, payload); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "nested.value") || strings.Contains(out, "other") {
		t.Fatalf("out=%s", out)
	}
}
