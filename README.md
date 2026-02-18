# sdo

`sdo` is a practical CLI wrapper for Scrape.do APIs.

## What it covers

- Sync API (`https://api.scrape.do`)
  - `GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS /` via `sdo scrape`
  - `GET /info` via `sdo info`
  - `GET /plugin/{path}` via `sdo plugin run`
- Async API (`https://q.scrape.do`, `X-Token` auth)
  - `POST /api/v1/jobs` via `sdo async submit`
  - `GET /api/v1/jobs` via `sdo async list`
  - `GET /api/v1/jobs/{job_id}` via `sdo async status`
  - `GET /api/v1/jobs/{job_id}/{task_id}` via `sdo async task`
  - `DELETE /api/v1/jobs/{job_id}` via `sdo async cancel`
  - `GET /api/v1/me` via `sdo async me`

## Install

```bash
make build
./bin/sdo --help
```

## Auth

Token resolution priority:

1. `--token`
2. `SCRAPEDO_TOKEN`
3. Config key `token`

Set token once:

```bash
sdo config set token YOUR_TOKEN
```

## Quick examples

Sync scrape:

```bash
sdo scrape https://httpbin.co/anything --geo us
sdo scrape https://httpbin.co/anything --render --wait-until domcontentloaded
sdo scrape https://httpbin.co/anything --method POST --body '{"hello":"world"}' --content-type application/json
sdo scrape https://example.com --render --width 1366 --height 768
sdo scrape https://example.com --play-with-browser '[{"Action":"WaitSelector","WaitSelector":"body"}]' --return-json
```

Target headers/cookies:

```bash
sdo scrape https://httpbin.co/anything --header 'User-Agent: my-agent' --custom-headers
sdo scrape https://httpbin.co/anything --set-cookie session=abc123
```

Plugin run:

```bash
sdo plugin run amazon/pdp \
  --url https://www.amazon.com/dp/B08N5WRWNW \
  --param asin=B08N5WRWNW \
  --param geocode=us \
  --param countryName='United States'
```

Async:

```bash
sdo async me --json
sdo async submit https://example.com --render
sdo async submit https://example.com --render --webhook-url https://example.com/callback --webhook-header 'Authorization: Bearer x'
sdo async list --page 1 --page-size 20
sdo async status <job_id>
sdo async task <job_id> <task_id>
sdo async wait <job_id> --wait-timeout 3m --interval 3s
sdo async cancel <job_id>
```

Discoverability:

```bash
sdo docs
sdo docs --json
sdo <command> --help
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
- `SCRAPEDO_CONFIG_DIR` (optional override for config directory)

## Async host override

If your network cannot resolve the default async host:

```bash
sdo --async-base-url https://your-async-host async status <job_id>
export SCRAPEDO_ASYNC_BASE_URL=https://your-async-host
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
