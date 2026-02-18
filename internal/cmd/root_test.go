package cmd

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestParseDurationOrMS(t *testing.T) {
	got, err := parseDurationOrMS("1500")
	if err != nil {
		t.Fatalf("parseDurationOrMS: %v", err)
	}
	if got != 1500*time.Millisecond {
		t.Fatalf("got=%v", got)
	}
	got, err = parseDurationOrMS("2s")
	if err != nil {
		t.Fatalf("parseDurationOrMS: %v", err)
	}
	if got != 2*time.Second {
		t.Fatalf("got=%v", got)
	}
	if _, err := parseDurationOrMS("-2s"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSplitCommaList(t *testing.T) {
	out := splitCommaList(" a, b,,c ")
	if len(out) != 3 || out[0] != "a" || out[2] != "c" {
		t.Fatalf("out=%v", out)
	}
}

func TestResolveHelpers(t *testing.T) {
	t.Setenv("SCRAPEDO_TEST_BOOL", "true")
	t.Setenv("SCRAPEDO_TEST_STRING", "env")
	if got := resolveString(false, "", "SCRAPEDO_TEST_STRING", "cfg", "fallback"); got != "env" {
		t.Fatalf("resolveString=%q", got)
	}
	if got := resolveBool(false, false, "SCRAPEDO_TEST_BOOL", false, false); !got {
		t.Fatalf("resolveBool expected true")
	}
}

func TestStableExitCode(t *testing.T) {
	if got := ExitCode(stableExitCode(context.Canceled)); got != exitCodeCancelled {
		t.Fatalf("exit=%d", got)
	}
	dnsErr := &net.DNSError{Err: "no such host", Name: "async.scrape.do"}
	if got := ExitCode(stableExitCode(dnsErr)); got != exitCodeRetryable {
		t.Fatalf("exit=%d", got)
	}
	err := usage("bad")
	if got := ExitCode(stableExitCode(err)); got != 2 {
		t.Fatalf("exit=%d", got)
	}
	wrapped := stableExitCode(errors.New("x"))
	if ExitCode(wrapped) != 1 {
		t.Fatalf("exit=%d", ExitCode(wrapped))
	}
}
