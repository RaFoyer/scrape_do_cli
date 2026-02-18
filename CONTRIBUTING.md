# Contributing

Thanks for contributing to `sdo`.

## Before You Start

- Read the [Code of Conduct](CODE_OF_CONDUCT.md).
- For vulnerabilities, use [SECURITY.md](SECURITY.md) instead of opening a public issue.
- Open an issue before large feature work so we can align on scope.

## Local Development

Prerequisites:

- Go `1.23+`
- `make`
- `git`

Setup:

```bash
git clone https://github.com/RaFoyer/scrape_do_cli.git
cd scrape_do_cli
make build
./bin/sdo --help
```

## Common Commands

```bash
make fmt
make test
make build
```

## Project Layout

- `cmd/sdo/main.go`: thin CLI entrypoint.
- `internal/cmd`: command definitions and runtime wiring.
- `internal/client`: Scrape.do HTTP client + request/response mapping.
- `internal/config`: local config file handling.
- `internal/outfmt`: JSON/plain output behavior.
- `internal/errfmt`: user-facing error formatting.

## Test Expectations

For behavior changes:

1. Add/adjust unit tests.
2. Run `make test`.
3. Include at least one manual CLI verification command in the PR.

When touching API integration behavior, include sanitized command/output snippets from:

- `sdo info --json`
- affected command(s), ideally both success and failure paths.

## Commit and Branch Guidance

- Keep commits small and logical.
- Prefer Conventional Commit style:
  - `feat: ...`
  - `fix: ...`
  - `docs: ...`
  - `test: ...`
- Keep unrelated changes in separate commits/PRs.

## Pull Requests

Use the PR template and include:

- Problem statement
- Scope of change
- Test evidence (`make test` + manual checks)
- Docs updates (if command surface changed)

A PR is ready when:

- CI is green
- review comments are resolved
- no unrelated changes remain

## Working With Secrets

- Never commit API tokens or credentials.
- Redact sensitive values from logs, screenshots, and PR text.
- Prefer environment variables for local testing (`SCRAPEDO_TOKEN`).

## Good First Contributions

Good starter tasks usually include:

- command help clarity improvements
- test coverage additions
- docs/examples improvements
- exit-code and error-message consistency fixes
