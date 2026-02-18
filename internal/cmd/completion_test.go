package cmd

import (
	"strings"
	"testing"
)

func TestCompletionScript_Bash(t *testing.T) {
	s, err := completionScript("bash")
	if err != nil {
		t.Fatalf("completionScript: %v", err)
	}
	if !strings.Contains(s, "__complete") || !strings.Contains(s, "complete -F _sdo_complete sdo") {
		t.Fatalf("unexpected script: %q", s)
	}
}

func TestCompleteWords(t *testing.T) {
	items, err := completeWords(1, []string{"sdo", "sc"})
	if err != nil {
		t.Fatalf("completeWords: %v", err)
	}
	found := false
	for _, item := range items {
		if item == "scrape" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected scrape completion in %v", items)
	}
}
