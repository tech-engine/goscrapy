# AGENTS.md

Project-level guidance for AI coding agents and human contributors working on
**GoScrapy** (`github.com/tech-engine/goscrapy`).

The global `AGENTS.md` rules (commit-message style, formatting, etc.) still
apply; this file only adds what's specific to *this* repository.

## Project at a glance

- **Language:** Go (requires `go 1.26+`, see `go.mod`).
- **Type:** Web-scraping framework inspired by Python's Scrapy.
- **Module path:** `github.com/tech-engine/goscrapy`.
- **CLIs:** `goscrapy` and `gos` alias (both installed from `cmd/`).
- **Layout:**
  - `cmd/cli`, `cmd/gos`, `cmd/goscrapy` — CLI entry points (Cobra-based).
  - `pkg/engine` — request/response loop, scheduler, worker pool glue.
  - `pkg/scheduler`, `pkg/worker`, `pkg/executor`, `pkg/executor_adapters` — execution pipeline.
  - `pkg/signal` — strongly-typed signal bus (`OnEngineStarted`, `OnItemScraped`, …).
  - `pkg/middlewaremanager`, `pkg/builtin/middlewares` — middleware system.
  - `pkg/pipeline_manager`, `pkg/builtin/pipelines/*` — export pipelines (CSV, JSON, MongoDB, Google Sheets, Firebase).
  - `pkg/gos` — public API surface (`gos.New`, `gosm` declarative mapping).
  - `pkg/telemetry`, `pkg/tui` — metrics + Bubble Tea terminal UI.
  - `pkg/logger`, `pkg/utils`, `pkg/core` — supporting infrastructure.
  - `_examples/*` — runnable example projects, each its own Go module.
  - `docs/` — architecture and telemetry diagrams.
  - `assets/` — logo, demo gif, sponsor banners (e.g. `nodemaven.svg`).

## Build, test, lint

Use the existing `Makefile` targets. They run with the race detector.

| Command | Purpose |
| :--- | :--- |
| `make test` | Run all unit tests. |
| `make test-race` | Run all tests with `-race`. |
| `make test-verbose` | Verbose run with `-race -v`. |
| `make test-pkg PKG=./pkg/scheduler` | Test a single package. |
| `make cover` | Total coverage summary. |
| `make cover-func` | Per-function coverage. |
| `make cover-html` | Open `coverage.html` in a browser. |
| `make tidy-examples` | `go mod tidy` across all `_examples`. |
| `make update-examples` | Bump `goscrapy` dep across all `_examples`. |

Useful ad-hoc commands:

```sh
go build ./...
go vet ./...
gofmt -l .
```

## Coding conventions

- **Go style:** follow `gofmt` + `go vet`. Run them before opening a PR.
- **Public API surface** lives in `pkg/gos` and `pkg/core`. Changes here are
  breaking — call them out clearly in the PR description.
- **Strongly-typed signals** are preferred. When adding a new signal, register
  it on the bus and expose a typed `On*` builder on `pkg/gos`.
- **Selectors:** `gosm` (`gos:"…"`, `gos_css:"…"`, `gos_xpath:"…"`) is the
  preferred way to declare extraction; avoid ad-hoc string parsing in spiders.
- **Pipelines** must be safe for concurrent calls. Use `engine.ErrDropItem` to
  drop an item explicitly.
- **Middlewares** must implement the `middlewaremanager.IMiddleware` interface
  and must be safe to register per-host.
- **No silent API breaks** during `v0.x` is fine, but document them in the
  PR title prefix `chore(api)!:` or `feat(api)!:`.

## Testing guidelines

- Unit tests live next to the package: `pkg/<name>/<file>_test.go`.
- E2E / example-driven tests live in `_examples/*` and may use a real local
  HTTP fixture or public test endpoints (e.g. `httpbin.org`, `books.toscrape.com`).
- Avoid network-dependent unit tests; mock or use `httptest.Server`.
- Always run `make test-race` before submitting — concurrency regressions are
  the #1 source of bugs in this codebase.

## Commit messages

The global `AGENTS.md` rule applies. Summary:

- Subject ≤ 50 chars, imperative mood, no trailing period.
- Body wrapped at 72 chars, only when it adds *useful* information.
- No raw diff in the message body.

## Pull request checklist

- [ ] `make test-race` passes locally.
- [ ] `go vet ./...` and `gofmt -l .` produce no output.
- [ ] Public API changes are called out in the PR description.
- [ ] New behaviour is covered by a unit or example test.
- [ ] Docs updated: `README.md`, `docs/architecture.md`, and/or in-code godoc.
- [ ] New sponsor / partner assets (SVG/PNG) live under `assets/` and are
      referenced from `README.md` via a relative path, not an external hotlink.

## Out of scope for agents

- Do **not** edit `LICENSE`.
- Do **not** bump the major Go version in `go.mod` without maintainer sign-off.
- Do **not** rewrite generated example projects under `_examples/*` as a "drive-by".
- Do **not** add new top-level dependencies without justifying the choice in
  the PR description (existing deps already cover HTTP, TUI, OAuth, Mongo,
  GSheets, Firebase, and TLS fingerprinting).
- Do **not** reorder the README sponsor tiers or add/remove sponsor entries
  without maintainer approval — these reflect active business agreements.

## Project links

- Repo: `github.com/tech-engine/goscrapy`
- Discord: https://discord.gg/FPvxETjYPH
- Wiki: https://github.com/tech-engine/goscrapy/wiki

## Sponsored by

### Platinum Sponsors:

This project is supported by [NodeMaven](https://go.nodemaven.com/techengineagentmd) — high quality Residential, Mobile, and ISP proxies for web scraping and automation. Use `GOSCRAPY35` for 35% off Residential/Mobile proxies and `GOSCRAPY40` for 40% off ISP proxies.
