# sdo

Scrape.do CLI wrapper inspired by gogcli architecture.

## Features

- Sync scrape command for `api.scrape.do`
- Async submit/status/wait commands for async jobs
- Generic plugin endpoint runner
- Account usage info command (`/info`)
- Local config management
- JSON/plain output modes with JSON field selection
- Shell completion and command schema output

## Install

Build locally:

```bash
make build
./bin/sdo --help
```

## Auth

Token resolution priority:

1. CLI flag `--token`
2. Env var `SCRAPEDO_TOKEN`
3. Config file key `token`

Set token once:

```bash
sdo config set token YOUR_TOKEN
```

## Quick examples

Sync scrape:

```bash
sdo scrape https://httpbin.co/anything --render --super --geo us
```

Async submit + wait:

```bash
sdo async submit https://example.com --render
sdo async wait <job_id>
sdo async wait <job_id> --wait-timeout 3m --interval 3s
```

Plugin run:

```bash
sdo plugin run amazon/pdp --url https://www.amazon.com/dp/B08N5WRWNW --param country=us
```

Info:

```bash
sdo info
```

Machine output:

```bash
sdo --json scrape https://example.com --results-only
sdo --json scrape https://example.com --select status_code,sdo_headers.scrape.do-request-cost
```

## Config keys

- `token`
- `base_url`
- `async_base_url`
- `timeout_ms`
- `default_output` (`json` or `plain`)

## Environment variables

- `SCRAPEDO_TOKEN`
- `SCRAPEDO_BASE_URL`
- `SCRAPEDO_ASYNC_BASE_URL`
- `SCRAPEDO_TIMEOUT`
- `SCRAPEDO_JSON`
- `SCRAPEDO_PLAIN`

## Async host override

If `async.scrape.do` cannot be resolved in your network:

```bash
sdo --async-base-url https://your-async-host async status <job_id>
```

or set:

```bash
export SCRAPEDO_ASYNC_BASE_URL=https://your-async-host
```
