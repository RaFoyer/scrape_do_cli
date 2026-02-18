package config

import (
	"path/filepath"
	"testing"
)

func TestSetAndGetValue(t *testing.T) {
	cfg := File{}
	if err := SetValue(&cfg, KeyToken, "abc"); err != nil {
		t.Fatalf("SetValue token: %v", err)
	}
	if got := GetValue(cfg, KeyToken); got != "abc" {
		t.Fatalf("token=%q", got)
	}
	if err := SetValue(&cfg, KeyTimeoutMS, "1200"); err != nil {
		t.Fatalf("SetValue timeout: %v", err)
	}
	if got := GetValue(cfg, KeyTimeoutMS); got != "1200" {
		t.Fatalf("timeout=%q", got)
	}
	if err := SetValue(&cfg, KeyDefaultOutput, "json"); err != nil {
		t.Fatalf("SetValue default_output: %v", err)
	}
	if got := GetValue(cfg, KeyDefaultOutput); got != "json" {
		t.Fatalf("default_output=%q", got)
	}
}

func TestSetValue_Validation(t *testing.T) {
	cfg := File{}
	if err := SetValue(&cfg, KeyTimeoutMS, "-1"); err == nil {
		t.Fatalf("expected timeout error")
	}
	if err := SetValue(&cfg, KeyDefaultOutput, "xml"); err == nil {
		t.Fatalf("expected default_output error")
	}
}

func TestParseKey(t *testing.T) {
	cases := map[string]Key{
		"token":          KeyToken,
		"base-url":       KeyBaseURL,
		"async_base_url": KeyAsyncBaseURL,
		"timeout":        KeyTimeoutMS,
		"default output": KeyDefaultOutput,
	}
	for in, want := range cases {
		got, err := ParseKey(in)
		if err != nil {
			t.Fatalf("ParseKey(%q): %v", in, err)
		}
		if got != want {
			t.Fatalf("ParseKey(%q)=%v want %v", in, got, want)
		}
	}
}

func TestReadWriteConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := File{Token: "abc", BaseURL: "https://api.scrape.do", TimeoutMS: 1000}
	if err := WriteConfig(cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}
	read, err := ReadConfig()
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}
	if read.Token != cfg.Token || read.BaseURL != cfg.BaseURL || read.TimeoutMS != cfg.TimeoutMS {
		t.Fatalf("read=%+v want=%+v", read, cfg)
	}
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}
	if filepath.Base(path) != "config.json" {
		t.Fatalf("path=%s", path)
	}
}
