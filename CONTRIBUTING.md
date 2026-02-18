# Contributing

Thanks for contributing to `sdo`.

## Development Setup

1. Install Go 1.22+.
2. Clone the repo.
3. Run tests.

```bash
make test
```

4. Build locally.

```bash
make build
./bin/sdo --help
```

## Branches and Commits

- Use short, focused branches.
- Keep commits logical and scoped.
- Use clear commit messages (for example: `feat: add async status polling`).

## Pull Requests

- Include a clear problem statement and solution summary.
- Add or update tests for behavior changes.
- Update docs/README when flags or commands change.
- Keep PRs focused; split unrelated work.

## Style

- Run formatting before opening a PR:

```bash
make fmt
```

- Ensure tests pass:

```bash
make test
```

## Reporting Issues

Use the issue templates and include:

- Reproduction steps
- Expected vs actual behavior
- `sdo version`
- Environment details (OS, shell, Go version)
