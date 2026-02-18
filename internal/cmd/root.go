package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kong"

	"github.com/ra/scrape_do_cli/internal/config"
	"github.com/ra/scrape_do_cli/internal/errfmt"
	"github.com/ra/scrape_do_cli/internal/outfmt"
)

type RootFlags struct {
	JSON         bool   `help:"Output JSON to stdout" aliases:"machine" short:"j"`
	Plain        bool   `help:"Output stable parseable text to stdout" short:"p"`
	ResultsOnly  bool   `name:"results-only" help:"In JSON mode, emit only primary result"`
	Select       string `name:"select" aliases:"pick,project" help:"In JSON mode, select comma-separated fields"`
	Token        string `help:"Scrape.do API token" env:"SCRAPEDO_TOKEN"`
	BaseURL      string `name:"base-url" help:"Sync API base URL" env:"SCRAPEDO_BASE_URL"`
	AsyncBaseURL string `name:"async-base-url" help:"Async API base URL" env:"SCRAPEDO_ASYNC_BASE_URL"`
	Timeout      string `help:"HTTP timeout duration (e.g. 30s, 2m)" env:"SCRAPEDO_TIMEOUT" default:"30s"`
	Verbose      bool   `help:"Enable verbose logging" short:"v"`
}

type CLI struct {
	RootFlags `embed:""`

	Version kong.VersionFlag `help:"Print version and exit"`

	Scrape     ScrapeCmd             `cmd:"" help:"Scrape a URL via sync API"`
	Async      AsyncCmd              `cmd:"" help:"Asynchronous scrape API commands"`
	Plugin     PluginCmd             `cmd:"" help:"Plugin endpoint commands"`
	Info       InfoCmd               `cmd:"" help:"Get account usage statistics"`
	Config     ConfigCmd             `cmd:"" help:"Manage local configuration"`
	VersionCmd VersionCmd            `cmd:"" name:"version" help:"Print version"`
	Completion CompletionCmd         `cmd:"" help:"Generate shell completion scripts"`
	Complete   CompletionInternalCmd `cmd:"" name:"__complete" hidden:"" help:"Internal completion helper"`
	Schema     SchemaCmd             `cmd:"" help:"Print machine-readable command schema"`
}

type exitPanic struct{ code int }

func Execute(args []string) (err error) {
	parser, cli, err := newParser(helpDescription())
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			if ep, ok := r.(exitPanic); ok {
				if ep.code == 0 {
					err = nil
					return
				}
				err = &ExitError{Code: ep.code, Err: errors.New("exited")}
				return
			}
			panic(r)
		}
	}()

	kctx, err := parser.Parse(args)
	if err != nil {
		parsedErr := wrapParseError(err)
		_, _ = fmt.Fprintln(os.Stderr, errfmt.Format(parsedErr))
		return parsedErr
	}

	cfg, err := config.ReadConfig()
	if err != nil {
		wrapped := &ExitError{Code: exitCodeConfig, Err: fmt.Errorf("read config: %w", err)}
		_, _ = fmt.Fprintln(os.Stderr, errfmt.Format(wrapped))
		return wrapped
	}

	runtime, err := resolveRuntime(kctx, &cli.RootFlags, cfg)
	if err != nil {
		wrapped := stableExitCode(err)
		_, _ = fmt.Fprintln(os.Stderr, errfmt.Format(wrapped))
		return wrapped
	}

	logLevel := slog.LevelWarn
	if runtime.Verbose {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	mode, err := outfmt.FromFlags(runtime.JSON, runtime.Plain)
	if err != nil {
		wrapped := &ExitError{Code: 2, Err: err}
		_, _ = fmt.Fprintln(os.Stderr, errfmt.Format(wrapped))
		return wrapped
	}

	ctx := context.Background()
	ctx = outfmt.WithMode(ctx, mode)
	ctx = outfmt.WithJSONTransform(ctx, outfmt.JSONTransform{ResultsOnly: runtime.ResultsOnly, Select: splitCommaList(runtime.Select)})
	ctx = withRuntimeOptions(ctx, runtime)

	kctx.Bind(kctx)
	kctx.Bind(&cli.RootFlags)
	kctx.BindTo(ctx, (*context.Context)(nil))

	err = kctx.Run()
	if err == nil {
		return nil
	}
	if ExitCode(err) == 0 {
		return nil
	}

	wrapped := stableExitCode(err)
	msg := strings.TrimSpace(errfmt.Format(wrapped))
	if msg != "" {
		_, _ = fmt.Fprintln(os.Stderr, msg)
	}
	return wrapped
}

func wrapParseError(err error) error {
	if err == nil {
		return nil
	}
	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		return &ExitError{Code: 2, Err: parseErr}
	}
	return err
}

type runtimeOptions struct {
	Token        string
	BaseURL      string
	AsyncBaseURL string
	Timeout      time.Duration
	JSON         bool
	Plain        bool
	ResultsOnly  bool
	Select       string
	Verbose      bool
}

func resolveRuntime(kctx *kong.Context, flags *RootFlags, cfg config.File) (runtimeOptions, error) {
	if flags == nil {
		flags = &RootFlags{}
	}

	runtime := runtimeOptions{
		Token:        resolveString(flagProvided(kctx, "token"), flags.Token, "SCRAPEDO_TOKEN", cfg.Token, ""),
		BaseURL:      resolveString(flagProvided(kctx, "base-url"), flags.BaseURL, "SCRAPEDO_BASE_URL", cfg.BaseURL, "https://api.scrape.do"),
		AsyncBaseURL: resolveString(flagProvided(kctx, "async-base-url"), flags.AsyncBaseURL, "SCRAPEDO_ASYNC_BASE_URL", cfg.AsyncBaseURL, "https://async.scrape.do"),
		JSON:         resolveBool(flagProvided(kctx, "json"), flags.JSON, "SCRAPEDO_JSON", cfg.DefaultOutput == "json", false),
		Plain:        resolveBool(flagProvided(kctx, "plain"), flags.Plain, "SCRAPEDO_PLAIN", cfg.DefaultOutput == "plain", false),
		ResultsOnly:  flags.ResultsOnly,
		Select:       flags.Select,
		Verbose:      flags.Verbose,
	}

	timeoutRaw := resolveString(flagProvided(kctx, "timeout"), flags.Timeout, "SCRAPEDO_TIMEOUT", timeoutFromConfig(cfg), "30s")
	d, err := parseDurationOrMS(timeoutRaw)
	if err != nil {
		return runtimeOptions{}, usagef("invalid --timeout value %q: %v", timeoutRaw, err)
	}
	runtime.Timeout = d

	return runtime, nil
}

func timeoutFromConfig(cfg config.File) string {
	if cfg.TimeoutMS <= 0 {
		return ""
	}
	return fmt.Sprintf("%dms", cfg.TimeoutMS)
}

func resolveString(flagSet bool, flagValue string, envKey string, cfgValue string, fallback string) string {
	if flagSet {
		return strings.TrimSpace(flagValue)
	}
	if envVal := strings.TrimSpace(os.Getenv(envKey)); envVal != "" {
		return envVal
	}
	if strings.TrimSpace(cfgValue) != "" {
		return strings.TrimSpace(cfgValue)
	}
	return fallback
}

func resolveBool(flagSet bool, flagValue bool, envKey string, cfgValue bool, fallback bool) bool {
	if flagSet {
		return flagValue
	}
	if envRaw := strings.TrimSpace(strings.ToLower(os.Getenv(envKey))); envRaw != "" {
		switch envRaw {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		default:
			return fallback
		}
	}
	if cfgValue {
		return true
	}
	return fallback
}

func parseDurationOrMS(v string) (time.Duration, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 30 * time.Second, nil
	}
	if strings.IndexFunc(v, func(r rune) bool { return r < '0' || r > '9' }) == -1 {
		ms, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		if ms < 0 {
			return 0, fmt.Errorf("must be non-negative")
		}
		return time.Duration(ms) * time.Millisecond, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, err
	}
	if d < 0 {
		return 0, fmt.Errorf("must be non-negative")
	}
	return d, nil
}

func splitCommaList(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func baseDescription() string {
	return "Scrape.do CLI"
}

func helpDescription() string {
	desc := baseDescription()
	cfgPath, err := config.ConfigPath()
	if err != nil {
		cfgPath = "unknown"
	}
	return fmt.Sprintf("%s\n\nConfig:\n  file: %s", desc, cfgPath)
}

func newParser(description string) (*kong.Kong, *CLI, error) {
	envMode := outfmt.FromEnv()
	vars := kong.Vars{
		"json":    strconv.FormatBool(envMode.JSON),
		"plain":   strconv.FormatBool(envMode.Plain),
		"version": VersionString(),
	}
	cli := &CLI{}
	parser, err := kong.New(
		cli,
		kong.Name("sdo"),
		kong.Description(description),
		kong.Vars(vars),
		kong.Writers(os.Stdout, os.Stderr),
		kong.Exit(func(code int) { panic(exitPanic{code: code}) }),
	)
	if err != nil {
		return nil, nil, err
	}
	return parser, cli, nil
}
