# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Purpose

A reference monorepo demonstrating production-grade CI/CD practices on GitHub Actions. The Go + chi API and Vue 3 frontend are concrete examples — the real output is the CI/CD patterns, not the applications themselves.

## Repository Status

Implementation is in progress via `feature/foundation` branch (worktree at `.worktrees/foundation`). The `main` branch holds only docs and configuration scaffolding. Active code work happens in the worktree.

## Architecture

Two independent services under `services/`, each with its own module/package system and Taskfile. A root `Taskfile.yml` aggregates them via `includes:` namespacing.

- **`services/api/`** — Go 1.23 + chi v5. Module path: `github.com/loic-vermeire/dx-connect-ci-scaffold/services/api`. Internal packages at `internal/store/` and `internal/handler/`. Entrypoint at `cmd/server/main.go`.
- **`services/web/`** — Vue 3 + Vite 6 + Vitest 3. Built static assets served by nginx:alpine in production.
- **`deploy/azure/`** — Azure Container Apps configs (placeholder for now).

## Commands

### From repo root (once Taskfiles are in place)
```bash
task api:build     # go build ./...
task api:test      # go test ./...
task api:lint      # golangci-lint run
task api:audit     # govulncheck ./...
task web:build     # npm run build
task web:test      # npm run test
task web:lint      # npm run lint
task web:audit     # npm audit
task build         # build both services
task test          # test both services
task up            # docker compose up --build
```

### From service directories
```bash
cd services/api && task test
cd services/web && task test
```

### Single test (Go)
```bash
cd services/api && go test ./internal/store/... -run TestItemStoreName -v
```

### Single test (Vue/Vitest)
```bash
cd services/web && npx vitest run src/components/ItemList.test.js
```

## Design Decisions

See `docs/superpowers/specs/2026-04-09-cicd-scaffold-design.md` for the full approved design. Key decisions:

- **`push-image` boolean input** on reusable workflows — `github.event_name` is unreliable inside `workflow_call` (always resolves to `workflow_call`). The dispatcher passes the boolean explicitly.
- **`docker/metadata-action`** handles `latest` automatically on semver — no explicit `type=raw,value=latest` needed.
- **release-please single-package** (`release-type: simple`) — structured for per-service migration by splitting the `.` entry in `release-please-config.json` and `.release-please-manifest.json`.
- **Trivy severity is graduated**: info on PR, HIGH/CRIT blocks on main, MEDIUM+ blocks on release.
- **Renovate** (not Dependabot) with `pinDigests: true` for Dockerfiles.

## Implementation Plans

- `docs/superpowers/plans/2026-04-09-foundation.md` — Plan A: scaffold, services, Taskfiles, Dockerfiles
- Plan B (CI/CD pipelines) to be written after Plan A completes.

## Git Worktrees

Active worktree: `.worktrees/foundation` on branch `feature/foundation`. The `.worktrees/` directory is gitignored.

When dispatching subagents for implementation tasks, always instruct them to `cd` to the worktree path once at the start and use relative paths for all subsequent commands. This avoids repeated permission prompts for chained `cd && command` patterns.
