# Endpoint Coverage

This document maps `sdo` commands to the Scrape.do API surface.

## Sync API (`https://api.scrape.do`)

- `sdo scrape <url>`
  - Endpoint: `GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS /`
  - Core options: proxy/geo/session/device, render controls, headers/cookies, retry/timeouts, output modes, `pureCookies`.
- `sdo info`
  - Endpoint: `GET /info`
- `sdo plugin run <plugin_path> [--url <url>]`
  - Endpoint: `GET /plugin/{plugin_path}`
  - `--url` is optional because some plugin endpoints are parameter-only.

## Async API (`https://q.scrape.do`)

- `sdo async submit [<url> ...]`
  - Endpoint: `POST /api/v1/jobs`
- `sdo async list`
  - Endpoint: `GET /api/v1/jobs`
- `sdo async status <job_id>`
  - Endpoint: `GET /api/v1/jobs/{job_id}`
- `sdo async task <job_id> <task_id>`
  - Endpoint: `GET /api/v1/jobs/{job_id}/{task_id}`
- `sdo async cancel <job_id>`
  - Endpoint: `DELETE /api/v1/jobs/{job_id}`
- `sdo async me`
  - Endpoint: `GET /api/v1/me`
- `sdo async wait <job_id>`
  - Polls `GET /api/v1/jobs/{job_id}` until terminal state.

## Output and Discovery Commands

- `sdo docs`
- `sdo schema [command path]`
- `sdo completion <bash|zsh|fish|powershell>`
- `sdo version`

## Research Note: Scrape.do Recursive Guard

As of **2026-02-18**, Scrape.do blocks scraping `*.scrape.do` with a recursive-request error.

Observed behavior:

- `sdo scrape https://docs.scrape.do ...` returns:
  - `Ooops :) You are sending recursive request for Scrape.do. Please check your target URL.`
- async jobs targeting `https://docs.scrape.do` complete at job level but task status becomes `error` with the same message.

Implication for maintainers:

- Do not depend on Scrape.do itself to fetch Scrape.do docs during automation.
- For endpoint/parameter verification, use primary sources directly:
  - Scrape.do docs pages (browser/manual)
  - Scrape.do official client repositories (e.g. `scrape-do/node-client`)
  - live endpoint checks against non-`scrape.do` targets.
