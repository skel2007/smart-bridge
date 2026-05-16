# AGENTS.md

Repo-specific instructions for agents working on `smart-bridge`.

## Project Shape

- Go module: `github.com/skel2007/smart-bridge`.
- CLI binary: `cmd/smart-bridge`.
- Keep `cmd/smart-bridge/main.go` as a thin entrypoint.
- Keep `internal/cli` as the CLI adapter layer.
- Do not introduce `pkg` until the repository actually exposes a public library API.

## Architecture

Current dependency direction:

```text
cmd/smart-bridge -> internal/cli -> internal/config, internal/tuya, internal/devices
```

Expected future HTTP app dependency direction:

```text
cmd/smart-bridge-http -> internal/api -> internal/config, internal/tuya, internal/devices
```

Rules:

- Shared behavior should live below adapter layers, not inside the CLI package.
- Prefer small, focused changes that match existing package boundaries.

## Workflow

- Work step by step.
- Before starting the next functional step, propose it and wait for explicit approval.
- Split work into reviewable commits of reasonable size; avoid mixing unrelated decisions in one commit.
- After each code step, review the written code with a fresh eye, run the relevant checks, and proactively ask whether to commit the completed reviewable step.
- Run focused tests after behavior changes; use `go test ./...` when the change is broad or cheap.
- Do not commit local configuration files or secrets.
