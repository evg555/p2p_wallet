# Repository Guidelines

## Project Structure & Module Organization
`cmd/p2p_wallet/main.go` is the application entrypoint. Core code lives under `internal/`: `domain` holds business entities and rules, `service` contains use cases, `repository` implements persistence, `handler` maps HTTP requests to services, and `infra` contains server, cache, and Postgres wiring. Generated OpenAPI types and server interfaces live in `internal/api`. Integration-style API tests are in `tests/`, SQL migrations are in `migration/`, and the source OpenAPI spec is in `spec/openapi/`.

## Build, Test, and Development Commands
Use `task` as the main interface for local work.

- `task build` builds `./bin/p2p_wallet`.
- `task run` builds and starts the app on the host.
- `task docker-run` rebuilds and starts the service in Docker.
- `task migrate-up` / `task migrate-down` apply or roll back Goose migrations.
- `task unit-test` runs unit tests for `internal/service`, `internal/domain`, and `internal/infra`.
- `task integration-test` runs repository tests.
- `task api-test` runs contract-style tests in `./tests`.
- `task lint` runs `golangci-lint`.
- `task check` runs generated-code checks, lint, tests, and build as a full pre-PR gate.

## Coding Style & Naming Conventions
Follow standard Go formatting with `gofmt` and `goimports`; both are enforced by `golangci-lint`. Use tabs as produced by `gofmt`, `CamelCase` for exported names, and short lowercase package names. Keep layer boundaries intact: `internal/domain` must not depend on infra code, `handler` must not import `repository`, and `repository` must not import `service`.

## Testing Guidelines
Tests use Go’s `testing` package with `stretchr/testify`. Keep test files next to the code they verify and name them `*_test.go`. Prefer table-driven tests for pure logic and focused scenario names such as `invalid currency` or `wallet already exists`. Run `task check` before opening a PR; repository tests may require Docker/Testcontainers.

## Commit & Pull Request Guidelines
Recent commits use short, lowercase, action-first subjects such as `fixed bug with UTC` and `implemented wallet repo methods`. Keep commit messages concise and specific. For pull requests, include a brief description, note schema or API changes, link the issue when applicable, and mention the verification commands you ran. If HTTP behavior changes, include example requests or affected endpoints.

## Configuration & Generated Code
Copy `.env.example` to `.env` for local setup. Use `HTTP_HOST=localhost` and `DB_HOST=localhost` when running on the host; keep Docker defaults when using `docker compose`. Regenerate API code with `task gen-api` and mocks with `task gen-mocks`, then ensure generated files are committed.
