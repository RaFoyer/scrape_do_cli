package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const AppName = "sdo"

type File struct {
	Token         string `json:"token,omitempty"`
	BaseURL       string `json:"base_url,omitempty"`
	AsyncBaseURL  string `json:"async_base_url,omitempty"`
	TimeoutMS     int    `json:"timeout_ms,omitempty"`
	DefaultOutput string `json:"default_output,omitempty"`
}

type Key int

const (
	KeyToken Key = iota
	KeyBaseURL
	KeyAsyncBaseURL
	KeyTimeoutMS
	KeyDefaultOutput
)

var (
	errInvalidKey   = errors.New("invalid config key")
	errInvalidValue = errors.New("invalid config value")
)

func Dir() (string, error) {
	if v := strings.TrimSpace(os.Getenv("SCRAPEDO_CONFIG_DIR")); v != "" {
		return v, nil
	}
	if v := strings.TrimSpace(os.Getenv("SCRAPEDO_CONFIG_HOME")); v != "" {
		return filepath.Join(v, AppName), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(base, AppName), nil
}

func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}
	return dir, nil
}

func ConfigPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func ReadConfig() (File, error) {
	path, err := ConfigPath()
	if err != nil {
		return File{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return File{}, nil
		}
		return File{}, fmt.Errorf("read config: %w", err)
	}
	var cfg File
	if err := json.Unmarshal(b, &cfg); err != nil {
		return File{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func WriteConfig(cfg File) error {
	if _, err := EnsureDir(); err != nil {
		return err
	}
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	b = append(b, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("commit config: %w", err)
	}
	return nil
}

func ParseKey(s string) (Key, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	switch s {
	case "token":
		return KeyToken, nil
	case "base_url":
		return KeyBaseURL, nil
	case "async_base_url":
		return KeyAsyncBaseURL, nil
	case "timeout_ms", "timeout":
		return KeyTimeoutMS, nil
	case "default_output", "output":
		return KeyDefaultOutput, nil
	default:
		return 0, fmt.Errorf("%w: %q", errInvalidKey, s)
	}
}

func KeyNames() []string {
	return []string{"token", "base_url", "async_base_url", "timeout_ms", "default_output"}
}

func GetValue(cfg File, key Key) string {
	switch key {
	case KeyToken:
		return cfg.Token
	case KeyBaseURL:
		return cfg.BaseURL
	case KeyAsyncBaseURL:
		return cfg.AsyncBaseURL
	case KeyTimeoutMS:
		if cfg.TimeoutMS <= 0 {
			return ""
		}
		return fmt.Sprintf("%d", cfg.TimeoutMS)
	case KeyDefaultOutput:
		return cfg.DefaultOutput
	default:
		return ""
	}
}

func SetValue(cfg *File, key Key, value string) error {
	if cfg == nil {
		return fmt.Errorf("%w: nil config", errInvalidValue)
	}
	value = strings.TrimSpace(value)
	switch key {
	case KeyToken:
		cfg.Token = value
	case KeyBaseURL:
		cfg.BaseURL = value
	case KeyAsyncBaseURL:
		cfg.AsyncBaseURL = value
	case KeyTimeoutMS:
		if value == "" {
			cfg.TimeoutMS = 0
			return nil
		}
		timeout, err := strconv.Atoi(value)
		if err != nil || timeout < 0 {
			return fmt.Errorf("%w: timeout_ms must be a non-negative integer", errInvalidValue)
		}
		cfg.TimeoutMS = timeout
	case KeyDefaultOutput:
		if value != "" && value != "json" && value != "plain" {
			return fmt.Errorf("%w: default_output must be one of json|plain", errInvalidValue)
		}
		cfg.DefaultOutput = value
	default:
		return fmt.Errorf("%w: unknown key", errInvalidKey)
	}
	return nil
}

func UnsetValue(cfg *File, key Key) error {
	return SetValue(cfg, key, "")
}
