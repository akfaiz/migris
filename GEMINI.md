# Migris Project Instructions

Migris is a database migration library for Go, inspired by Laravel's migrations. It combines the power of [pressly/goose](https://github.com/pressly/goose) with a fluent schema builder.

## Project Overview

- **Purpose:** Provide a fluent, Laravel-like API for defining and managing database schemas in Go.
- **Main Technologies:** Go (1.25+), [goose](https://github.com/pressly/goose), MySQL, PostgreSQL (pgx), SQLite3.
- **Architecture:**
    - `schema/`: The core Schema Builder API. It uses a **Builder -> Blueprint -> Grammar -> DDL** pipeline.
        - `Blueprint`: Collects column and index definitions.
        - `Grammar`: Translates Blueprint definitions into dialect-specific SQL (DDL).
        - `Builder`: Orchestrates the creation and modification of tables.
    - `internal/`: Internal components including dialect management, logging, and configuration.
    - `extra/`: Integration helpers for popular CLI frameworks like `Cobra` and `urfave/cli`.
    - Root: High-level migration management (`Migrate`, `Registry`).

## Building and Running

The project uses a `Makefile` for common development tasks:

- **Install Dependencies:** `make install`
- **Format Code:** `make fmt`
- **Lint Code:** `make lint` or `make lint-fix`
- **Run Tests:** `make test` or `go test ./...`
- **Generate Coverage:** `make testcov` or `make testcov-html`

## Development Conventions

- **Code Style:** Use idiomatic Go and follow [Effective Go](https://golang.org/doc/effective_go.html). Run `go fmt` before submitting.
- **Testing:** 
    - ALWAYS add or update tests when making changes.
    - Grammar implementations (e.g., `mysql_grammar.go`) MUST have corresponding tests (e.g., `mysql_grammar_test.go`) to verify DDL generation.
    - Use `github.com/stretchr/testify` for assertions.
- **Grammar Implementation:**
    - When adding new fluent modifiers to `Blueprint`, ensure they are implemented in all supported grammars (MySQL, Postgres, SQLite).
    - Maintain consistent modifier order in DDL generation to match established conventions and existing tests.
    - Standard MySQL modifier order: `Unsigned`, `Charset`, `Collate`, `VirtualAs`, `StoredAs`, `Nullable`, `Default`, `Increment`, `OnUpdate`, `Comment`, `After`, `First`.
- **Backward Compatibility:** Keep the public API stable. Avoid breaking changes unless absolutely necessary.
- **Transaction Safety:** Ensure all schema operations are performed within the provided `schema.Context` (which typically wraps a database transaction).

## Key Files

- `schema/blueprint.go`: The central place for defining the fluent API.
- `schema/grammar.go`: The interface for dialect-specific SQL generation.
- `schema/column_definition.go`: Defines the attributes and modifiers for table columns.
- `migrate.go`: The main entry point for running migrations.
- `internal/dialect/dialect.go`: Handles database dialect identification and mapping to Goose dialects.
