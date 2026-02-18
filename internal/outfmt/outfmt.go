package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Mode struct {
	JSON  bool
	Plain bool
}

type ParseError struct{ msg string }

func (e *ParseError) Error() string { return e.msg }

func FromFlags(jsonOut bool, plainOut bool) (Mode, error) {
	if jsonOut && plainOut {
		return Mode{}, &ParseError{msg: "invalid output mode (cannot combine --json and --plain)"}
	}
	return Mode{JSON: jsonOut, Plain: plainOut}, nil
}

func FromEnv() Mode {
	return Mode{
		JSON:  envBool("SCRAPEDO_JSON"),
		Plain: envBool("SCRAPEDO_PLAIN"),
	}
}

type modeKey struct{}

type JSONTransform struct {
	ResultsOnly bool
	Select      []string
}

type transformKey struct{}

func WithMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, modeKey{}, mode)
}

func FromContext(ctx context.Context) Mode {
	if v := ctx.Value(modeKey{}); v != nil {
		if m, ok := v.(Mode); ok {
			return m
		}
	}
	return Mode{}
}

func IsJSON(ctx context.Context) bool  { return FromContext(ctx).JSON }
func IsPlain(ctx context.Context) bool { return FromContext(ctx).Plain }

func WithJSONTransform(ctx context.Context, t JSONTransform) context.Context {
	return context.WithValue(ctx, transformKey{}, t)
}

func JSONTransformFromContext(ctx context.Context) (JSONTransform, bool) {
	v := ctx.Value(transformKey{})
	if v == nil {
		return JSONTransform{}, false
	}
	t, ok := v.(JSONTransform)
	return t, ok
}

func WriteJSON(ctx context.Context, w io.Writer, v any) error {
	if t, ok := JSONTransformFromContext(ctx); ok && (t.ResultsOnly || len(t.Select) > 0) {
		transformed, err := applyJSONTransform(v, t)
		if err != nil {
			return fmt.Errorf("transform json: %w", err)
		}
		v = transformed
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func applyJSONTransform(v any, t JSONTransform) (any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	var anyV any
	if err := json.Unmarshal(b, &anyV); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	if t.ResultsOnly {
		anyV = unwrapPrimary(anyV)
	}
	if len(t.Select) > 0 {
		anyV = selectFields(anyV, t.Select)
	}
	return anyV, nil
}

func unwrapPrimary(v any) any {
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}
	if r, ok := m["results"]; ok {
		return r
	}
	meta := map[string]struct{}{
		"nextPageToken": {},
		"next_cursor":   {},
		"has_more":      {},
		"count":         {},
		"query":         {},
		"dry_run":       {},
		"dryRun":        {},
		"op":            {},
		"action":        {},
		"note":          {},
		"notes":         {},
	}
	candidates := make([]string, 0, len(m))
	for k := range m {
		if _, ok := meta[k]; ok {
			continue
		}
		candidates = append(candidates, k)
	}
	if len(candidates) == 1 {
		return m[candidates[0]]
	}
	for _, k := range candidates {
		if _, ok := m[k].([]any); ok {
			return m[k]
		}
	}
	known := []string{"content", "data", "job", "response", "info"}
	for _, k := range known {
		if val, ok := m[k]; ok {
			return val
		}
	}
	return v
}

func selectFields(v any, fields []string) any {
	switch vv := v.(type) {
	case []any:
		out := make([]any, 0, len(vv))
		for _, it := range vv {
			out = append(out, selectFieldsFromItem(it, fields))
		}
		return out
	default:
		return selectFieldsFromItem(v, fields)
	}
}

func selectFieldsFromItem(v any, fields []string) any {
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}
	out := make(map[string]any, len(fields))
	for _, f := range fields {
		if val, ok := getAtPath(m, f); ok {
			out[f] = val
		}
	}
	return out
}

func getAtPath(v any, path string) (any, bool) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, false
	}
	segs := strings.Split(path, ".")
	cur := v
	for _, seg := range segs {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			return nil, false
		}
		switch c := cur.(type) {
		case map[string]any:
			next, ok := c[seg]
			if !ok {
				return nil, false
			}
			cur = next
		case []any:
			i, err := strconv.Atoi(seg)
			if err != nil || i < 0 || i >= len(c) {
				return nil, false
			}
			cur = c[i]
		default:
			return nil, false
		}
	}
	return cur, true
}

func envBool(key string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
