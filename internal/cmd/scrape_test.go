package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadJSONInput(t *testing.T) {
	t.Run("inline", func(t *testing.T) {
		out, err := readJSONInput(`[{"Action":"WaitSelector"}]`, "", "--play-with-browser", "--play-with-browser-file")
		if err != nil {
			t.Fatalf("readJSONInput: %v", err)
		}
		if out == "" {
			t.Fatalf("expected value")
		}
	})

	t.Run("file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "pwb.json")
		if err := os.WriteFile(path, []byte(`{"Action":"Click"}`), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		out, err := readJSONInput("", path, "--play-with-browser", "--play-with-browser-file")
		if err != nil {
			t.Fatalf("readJSONInput: %v", err)
		}
		if out == "" {
			t.Fatalf("expected value")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if _, err := readJSONInput("{", "", "--play-with-browser", "--play-with-browser-file"); err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestParseJSONAny(t *testing.T) {
	v, err := parseJSONAny(`[{"Action":"WaitSelector"}]`, "--play-with-browser")
	if err != nil {
		t.Fatalf("parseJSONAny: %v", err)
	}
	if _, ok := v.([]any); !ok {
		t.Fatalf("value=%T", v)
	}
	if _, err := parseJSONAny("{", "--play-with-browser"); err == nil {
		t.Fatalf("expected error")
	}
}
