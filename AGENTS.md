# Repository Guidelines

## Project Structure & Module Organization
Core migration APIs live at the repository root (`up.go`, `down.go`, `migrate.go`, `create.go`, `reset.go`, `status.go`).
Schema builders and SQL grammar implementations are in `schema/` (MySQL, PostgreSQL, SQLite).
Internal helper packages are in `internal/` (`config`, `dialect`, `logger`, `parser`, `util`).
Runnable examples are in `examples/` (`basic`, `migriscli`, `migriscobra`), while optional integrations are in `extra/`.
Tests are colocated with source files as `*_test.go`.

## Build, Test, and Development Commands
Use `make` targets:
- `make install`: download and tidy module dependencies.
- `make install-tools`: install `golangci-lint` and `gotestsum`.
- `make fmt`: run `go fmt ./...`.
- `make lint`: run strict lint checks (`golangci-lint run --timeout=2m --verbose`).
- `make test`: run all tests via `gotestsum`.
- `make testcov`: run tests with coverage output (`coverage.out`).

For quick checks, `go test ./...` is acceptable, but prefer `make test` in PR validation.

## Coding Style & Naming Conventions
Target Go `1.24` (see `go.mod`). Follow idiomatic Go and keep exported APIs stable.
Formatting is required before commit (`go fmt` and linter formatters like `goimports`/`golines` are enforced).
Use descriptive CamelCase for exported identifiers and short, clear lowerCamelCase for locals.
Migration filenames in examples follow timestamped snake case (for example, `20250904164848_create_users_table.go`).

## Testing Guidelines
Use Go’s `testing` package with `testify` assertions where helpful.
Name tests clearly with `TestXxx` and colocate them with implementation.
Run `make test` locally before opening a PR; use `make testcov` when changing migration logic, schema builders, or SQL grammar behavior.
Add regression tests for bug fixes.

## Commit & Pull Request Guidelines
Commit messages must follow Conventional Commits, enforced by `lefthook`:
`feat|fix|docs|style|refactor|test|chore(scope): summary`.
Examples from history include `feat: add sqlite3 support` and `fix: ...`.

Before pushing, run: `make fmt && make lint && make test`.
PRs should include:
- A concise change summary and motivation.
- Linked issue(s) when applicable.
- Notes on behavior changes and test coverage updates.
- Example CLI output or screenshots only when UX/output changed.
