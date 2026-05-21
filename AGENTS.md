# AGENTS.md

Repo-specific instructions for agents working on `smart-bridge`.

## Project Shape

- Go module: `github.com/skel2007/smart-bridge`.
- CLI binary: `cmd/smart-bridge`.
- HTTP server binary: `cmd/smart-bridge-server`.
- Keep `cmd/smart-bridge/main.go` as a thin entrypoint.
- Keep `cmd/smart-bridge-server/main.go` as a thin entrypoint.
- Keep `internal/cli` as the CLI adapter layer.
- Keep `internal/server` as the HTTP server wiring layer.
- Do not introduce `pkg` until the repository actually exposes a public library API.

## Architecture

CLI dependency direction:

```text
cmd/smart-bridge -> internal/cli -> internal/config, internal/tuya, internal/devices
```

HTTP server dependency direction:

```text
cmd/smart-bridge-server -> internal/server
internal/server -> internal/config, internal/tuya, internal/yandex
internal/yandex -> internal/devices
internal/tuya -> internal/devices
```

Rules:

- Shared behavior should live below adapter layers, not inside the CLI package.
- `cmd/*` entrypoints should parse runtime flags, build process-level dependencies such as loggers and signal contexts, and delegate.
- `internal/cli` owns CLI wiring: command structure, flag handling, output formatting, config validation for CLI commands, and platform gateway construction.
- `internal/server` owns HTTP service wiring: config validation, platform gateway construction, downstream handler construction, route mounting, and service lifecycle.
- Prefer small, focused changes that match existing package boundaries.
- Prefer established Go libraries for non-trivial infrastructure concerns. Keep tiny local code only when a dependency would not reduce complexity.
- Leave short comments only when a fresh reader cannot infer intent from names, tests, or docs; prefer comments for external API quirks, fallback behavior, concurrency constraints, and intentionally defensive logic.

## Workflow

- Work step by step.
- Before starting the next functional step, propose it and wait for explicit approval.
- Split work into reviewable commits of reasonable size; avoid mixing unrelated decisions in one commit.
- If the working copy is managed by jj, treat one jj change as the reviewable unit.
  Prefer `jj status`, `jj diff`, `jj describe`, `jj new`, and `jj squash`;
  use direct Git history commands only when the user asks for them.
- After each code-edit iteration, review the written code with a fresh eye and run the relevant checks before proposing a commit.
- Proactively ask whether to commit only after the latest review pass is complete and the diff is still a reasonable reviewable unit.
- After committing a reviewable step, summarize the commit briefly, propose the next small step, and wait for explicit approval before starting it.
- Run focused tests after behavior changes; use `go test ./...` when the change is broad or cheap.
- Do not commit local configuration files or secrets.
