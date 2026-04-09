# CI/CD Scaffold Monorepo — Design Spec

**Date:** 2026-04-09
**Status:** Approved

---

## Overview

A reference monorepo demonstrating production-grade CI/CD practices on GitHub Actions. The codebase consists of a minimal Go + chi backend API and a Vue 3 frontend — chosen as concrete targets for linting, testing, containerisation, and versioning patterns, not as the focus of the work itself.

The primary output is container images. Deployment is kept thin and swappable: currently targeting Azure Container Apps, with AKS as a likely future target.

---

## Repository Structure

```
dx-connect-ci-scaffold/
├── services/
│   ├── api/                        # Go + chi backend
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   ├── Dockerfile
│   │   ├── Taskfile.yml
│   │   └── go.mod
│   └── web/                        # Vue 3 frontend
│       ├── src/
│       ├── nginx.conf
│       ├── Dockerfile
│       ├── Taskfile.yml
│       └── package.json
├── deploy/
│   └── azure/                      # ACA container app configs
├── .github/
│   └── workflows/
│       ├── ci.yml                  # dispatcher
│       ├── _go-service.yml         # reusable Go CI
│       ├── _node-service.yml       # reusable Node/Vue CI
│       ├── release.yml             # release-please + stable image push
│       └── security.yml            # scheduled security scans
├── scripts/
│   └── bootstrap-github.sh         # branch protection via gh CLI
├── Taskfile.yml                    # root, includes both services
├── docker-compose.yml              # local dev
├── renovate.json
├── release-please-config.json
└── .release-please-manifest.json
```

---

## Task Runner (Taskfile)

Each service has its own `Taskfile.yml`. The root `Taskfile.yml` includes both using namespacing, so tasks are available from anywhere in the repo.

### Root Taskfile
```yaml
includes:
  api:
    taskfile: ./services/api/Taskfile.yml
    dir: ./services/api
  web:
    taskfile: ./services/web/Taskfile.yml
    dir: ./services/web

tasks:
  build: { deps: [api:build, web:build] }
  test:  { deps: [api:test,  web:test]  }
  lint:  { deps: [api:lint,  web:lint]  }
  up:    docker compose up --build
```

### Service Taskfiles (same shape for both)
```yaml
tasks:
  build:        # go build ./... / npm run build
  test:         # go test ./... / npm run test
  lint:         # golangci-lint run / npm run lint
  audit:        # govulncheck ./... / npm audit
  run:          # go run ./cmd/server / npm run dev
  docker:build: # docker build -t <name> .
  docker:run:   # docker run -p <port>:<port> <name>
```

A developer can `cd services/api && task test` or run `task api:test` from root. CI always invokes from root via reusable workflows.

---

## CI Workflow Architecture

### Pattern: Reusable workflows + dispatcher

```
.github/workflows/
├── ci.yml              # triggered on push/PR, detects changed paths
├── _go-service.yml     # workflow_call: lint, test, build, scan
├── _node-service.yml   # workflow_call: lint, test, build, scan
├── release.yml         # triggered on v* tags
└── security.yml        # scheduled weekly
```

Reusable workflows (`_*.yml`) accept inputs:
- `service-path` — e.g. `services/api`
- `image-name` — e.g. `ghcr.io/org/api`
- `push-image` — boolean, passed by the dispatcher based on branch/event context (true only for main and release tags)

Adding a new service means adding one entry in `ci.yml` and reusing an existing `_*.yml` template.

### Trigger Matrix

| Trigger | What runs | Image pushed? |
|---|---|---|
| Push to feature branch | dispatcher → reusable workflow(s) matching changed paths: lint + test + build + docker build + dep scan + trivy (info) + secret scan | No |
| PR to main | Same as above — results required to pass before merge | No |
| Push / merge to main | Same + trivy (HIGH/CRIT blocks) + push `sha-<hash>` and `edge` images | Yes (`edge`, `sha-`) |
| Release tag `v*` | release.yml: build + trivy (MEDIUM+ blocks) + push stable semver images | Yes (semver) |
| Weekly schedule | security.yml: full trivy scan + dep audit on latest images | No |

---

## Container Images

### Registry

`ghcr.io` (GitHub Container Registry) — no extra credentials required, integrated with GitHub Actions OIDC via `permissions: packages: write`.

### Image names
```
ghcr.io/<org>/api:<tag>
ghcr.io/<org>/web:<tag>
```

### Tagging (docker/metadata-action)

```yaml
tags: |
  type=semver,pattern={{version}}         # 1.2.3  (release only)
  type=semver,pattern={{major}}.{{minor}} # 1.2    (release only)
  type=semver,pattern={{major}}           # 1      (release only)
  type=sha                                # sha-abc1234 (every push to main)
  type=edge,branch=main                   # edge   (tracks main)
```

`latest` is applied automatically by `docker/metadata-action` on the highest semver tag — no explicit `type=raw` needed.

Feature branch builds: docker build runs but image is not pushed. The dispatcher passes `push-image: false` for feature branches and PRs, `push-image: true` for main and release tags. Inside the reusable workflow, `build-push-action` uses `push: ${{ inputs.push-image }}`. Note: using `github.event_name` inside a reusable workflow is unreliable — it always resolves to `workflow_call`.

### Dockerfiles

**api (Go) — distroless runtime:**
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/server /server
USER nonroot:nonroot
ENTRYPOINT ["/server"]
```

**web (Vue) — nginx runtime:**
```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

---

## Release Management

### release-please configuration

Single-package setup, structured so migration to per-service versioning (strategy B) is mechanical: split the `.` entry in both config and manifest into `services/api` and `services/web`.

**release-please-config.json:**
```json
{
  "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
  "bump-minor-pre-major": false,
  "packages": {
    ".": {
      "release-type": "simple",
      "changelog-path": "CHANGELOG.md"
    }
  },
  "changelog-sections": [
    { "type": "feat",     "section": "Features"      },
    { "type": "fix",      "section": "Bug Fixes"     },
    { "type": "perf",     "section": "Performance"   },
    { "type": "revert",   "section": "Reverts"       },
    { "type": "docs",     "section": "Documentation" },
    { "type": "chore",    "hidden": true              },
    { "type": "refactor", "hidden": true              },
    { "type": "test",     "hidden": true              },
    { "type": "ci",       "hidden": true              }
  ]
}
```

**.release-please-manifest.json:**
```json
{ ".": "0.1.0" }
```

### Release flow

1. Feature work lands on `main` via normal PRs using conventional commits
2. After each merge to `main`, release-please auto-creates/updates a release PR (version bump + CHANGELOG update)
3. The release PR accumulates unreleased changes — merge it when ready to ship
4. Merging the release PR creates a git tag (`v1.2.3`) → triggers `release.yml` → builds and pushes stable images → GitHub Release created with generated changelog

### Bumping to 1.0.0

`bump-minor-pre-major: false` is the default and is stated explicitly for clarity. A commit with `BREAKING CHANGE:` in the footer triggers a major bump — including across the `0.x → 1.0.0` boundary. Use this intentionally when the API is stable:

```
feat!: first stable release

BREAKING CHANGE: stable public API
```

---

## Security Gates

| Gate | Action | Runs on | Blocks? |
|---|---|---|---|
| Go lint | `golangci-lint-action@v6` | every push / PR | yes |
| Vue lint | `task web:lint` (eslint) | every push / PR | yes |
| Go vuln scan | `golang/govulncheck-action@v1` | every push / PR | yes |
| npm audit | `task web:audit` | every push / PR | yes |
| Secret scan | `trufflesecurity/trufflehog@v3` | every push / PR | yes |
| Image scan (trivy) | `aquasecurity/trivy-action@v0` | PR: info; main: HIGH/CRIT; release: MEDIUM+ | graduated |
| SARIF upload | `github/codeql-action/upload-sarif@v3` | main + release | no (visibility) |
| Full scheduled scan | `security.yml` | weekly | no (report only) |

---

## Dependency Updates (Renovate)

Renovate manages all dependency types including Docker base image digests.

**renovate.json:**
```json
{
  "extends": ["config:recommended"],
  "dockerfile": { "enabled": true },
  "pinDigests": true
}
```

Behaviour:
- Go modules, npm packages, and GitHub Actions receive auto-PRs
- Patch/minor updates are grouped into batched PRs
- Major updates are individual PRs for review
- Dockerfile `FROM` lines are pinned to digest (`sha256:...`) and auto-updated when the upstream image changes

---

## Branch Protection

Branch protection rules for `main` must be configured as GitHub repo settings — they cannot be committed as files. A `scripts/bootstrap-github.sh` script configures them via `gh api` so the intent is documented and reproducible.

Required status checks (jobs that must pass before merge):
- `lint` (api + web)
- `test` (api + web)
- `govulncheck`
- `npm-audit`
- `trufflehog`
- `trivy` (PR-mode, informational threshold)

Additional rules:
- Require PR before merging
- Require up-to-date branch before merge
- Do not allow force pushes

---

## Environments

One environment to start. Dev deploy targets Azure Container Apps using images tagged `edge` (latest from main) or `sha-<hash>` for pinned deploys. The `deploy/azure/` directory holds ACA container app configs.

The deploy step is intentionally thin — a `gh workflow dispatch` or `az containerapp update` call with the new image tag. This makes it straightforward to replace with a GitOps approach (Flux/ArgoCD) or migrate to AKS without changing the image-building side of the pipeline.

---

## Migration Notes

### Monorepo versioning A → B (per-service)

1. Update `release-please-config.json`: replace `"."` package with `"services/api"` and `"services/web"`, set appropriate `release-type` per service (`go` / `node`)
2. Update `.release-please-manifest.json`: split single entry into per-service entries
3. Update `release.yml`: key image builds off per-service tags (`api/v1.2.3`) instead of repo-level tags (`v1.2.3`)
4. Existing git tags remain valid; new per-service tags start fresh going forward
