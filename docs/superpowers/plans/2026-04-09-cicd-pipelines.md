# Plan B — CI/CD Pipelines

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire the monorepo to production-grade GitHub Actions CI/CD: reusable workflows, container image builds, release-please, Renovate, scheduled security scans, and branch protection.

**Architecture:**
- `ci.yml` dispatcher detects changed paths and calls reusable `_go-service.yml` / `_node-service.yml`
- Reusable workflows: lint → test → vuln scan → docker build(/push) → secret scan → trivy
- `push-image` boolean input controls whether images are pushed (true only on main push and release tags)
- `trivy-severity` input controls blocking threshold (empty = info only on PRs, HIGH,CRITICAL on main)
- `release.yml`: release-please PR management on push to main; semver image publish on `v*` tags
- `security.yml`: weekly scheduled full scans
- `renovate.json`: dependency updates with digest pinning
- `scripts/bootstrap-github.sh`: branch protection via `gh api`

**Registry:** `ghcr.io/lvermeire/<service>` — GITHUB_TOKEN, no extra secrets needed.

**Image tagging (docker/metadata-action):**
- Every push to main: `sha-<hash>`, `edge`
- Release tags: `1.2.3`, `1.2`, `1`, `latest` (auto by metadata-action on semver)

> **Note:** Action versions use semver tags here (e.g. `actions/checkout@v4`). Renovate will pin them to `sha256:` digests on its first run after `renovate.json` is in place.

---

## File Map

```
.github/
  workflows/
    ci.yml
    _go-service.yml
    _node-service.yml
    release.yml
    security.yml
renovate.json
release-please-config.json
.release-please-manifest.json
CHANGELOG.md
scripts/
  bootstrap-github.sh
```

---

## Task 1: Release Management Configs

**Files:**
- Create: `release-please-config.json`
- Create: `.release-please-manifest.json`
- Create: `CHANGELOG.md`

- [ ] **Step 1: Write release-please-config.json**

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

- [ ] **Step 2: Write .release-please-manifest.json**

```json
{ ".": "0.1.0" }
```

- [ ] **Step 3: Create empty CHANGELOG.md**

```markdown
# Changelog
```

- [ ] **Step 4: Commit**

```bash
git add release-please-config.json .release-please-manifest.json CHANGELOG.md
git commit -m "chore: add release-please config and manifest"
```

---

## Task 2: Renovate Config

**Files:**
- Create: `renovate.json`

- [ ] **Step 1: Write renovate.json**

```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": ["config:recommended"],
  "pinDigests": true,
  "packageRules": [
    {
      "matchUpdateTypes": ["minor", "patch"],
      "groupName": "minor and patch updates",
      "automerge": false
    }
  ]
}
```

- [ ] **Step 2: Commit**

```bash
git add renovate.json
git commit -m "chore: add Renovate config with digest pinning"
```

---

## Task 3: Reusable Go Service Workflow

**Files:**
- Create: `.github/workflows/_go-service.yml`

- [ ] **Step 1: Create .github/workflows/ directory**

```bash
mkdir -p .github/workflows
```

- [ ] **Step 2: Write _go-service.yml**

`.github/workflows/_go-service.yml`:

```yaml
name: Go Service CI

on:
  workflow_call:
    inputs:
      service-path:
        description: 'Path to the service root (e.g. services/api)'
        required: true
        type: string
      image-name:
        description: 'Full image name including registry (e.g. ghcr.io/org/api)'
        required: true
        type: string
      push-image:
        description: 'Push the built image to the registry'
        required: true
        type: boolean
      trivy-severity:
        description: 'Comma-separated severity levels that cause a non-zero exit. Empty = informational only.'
        required: false
        type: string
        default: ''

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: ${{ inputs.service-path }}/go.mod
          cache: true
      - uses: golangci/golangci-lint-action@v6
        with:
          working-directory: ${{ inputs.service-path }}

  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: ${{ inputs.service-path }}/go.mod
          cache: true
      - name: Run tests
        working-directory: ${{ inputs.service-path }}
        run: go test ./...

  govulncheck:
    name: govulncheck
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: ${{ inputs.service-path }}/go.mod
          cache: true
      - uses: golang/govulncheck-action@v1
        with:
          work-dir: ${{ inputs.service-path }}

  build-image:
    name: Build image
    runs-on: ubuntu-latest
    needs: [lint, test, govulncheck]
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        if: inputs.push-image
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Extract image metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ inputs.image-name }}
          tags: |
            type=sha
            type=edge,branch=main
      - uses: docker/build-push-action@v6
        with:
          context: ${{ inputs.service-path }}
          push: ${{ inputs.push-image }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  secret-scan:
    name: Secret scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: trufflesecurity/trufflehog@v3
        with:
          extra_args: --only-verified

  trivy:
    name: Trivy scan
    runs-on: ubuntu-latest
    needs: build-image
    permissions:
      security-events: write
    steps:
      - uses: actions/checkout@v4
      - name: Scan with Trivy
        uses: aquasecurity/trivy-action@0.28.0
        with:
          scan-type: fs
          scan-ref: ${{ inputs.service-path }}
          format: sarif
          output: trivy-results.sarif
          severity: ${{ inputs.trivy-severity != '' && inputs.trivy-severity || 'CRITICAL,HIGH,MEDIUM,LOW,UNKNOWN' }}
          exit-code: ${{ inputs.trivy-severity != '' && '1' || '0' }}
      - name: Upload SARIF
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-results.sarif
```

- [ ] **Step 3: Validate YAML syntax**

```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/_go-service.yml'))" && echo "Valid YAML"
```

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/_go-service.yml
git commit -m "ci: add reusable Go service workflow"
```

---

## Task 4: Reusable Node Service Workflow

**Files:**
- Create: `.github/workflows/_node-service.yml`

- [ ] **Step 1: Write _node-service.yml**

`.github/workflows/_node-service.yml`:

```yaml
name: Node Service CI

on:
  workflow_call:
    inputs:
      service-path:
        description: 'Path to the service root (e.g. services/web)'
        required: true
        type: string
      image-name:
        description: 'Full image name including registry (e.g. ghcr.io/org/web)'
        required: true
        type: string
      push-image:
        description: 'Push the built image to the registry'
        required: true
        type: boolean
      trivy-severity:
        description: 'Comma-separated severity levels that cause a non-zero exit. Empty = informational only.'
        required: false
        type: string
        default: ''

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version-file: ${{ inputs.service-path }}/package.json
          cache: npm
          cache-dependency-path: ${{ inputs.service-path }}/package-lock.json
      - name: Install dependencies
        working-directory: ${{ inputs.service-path }}
        run: npm ci
      - name: Lint
        working-directory: ${{ inputs.service-path }}
        run: npm run lint

  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version-file: ${{ inputs.service-path }}/package.json
          cache: npm
          cache-dependency-path: ${{ inputs.service-path }}/package-lock.json
      - name: Install dependencies
        working-directory: ${{ inputs.service-path }}
        run: npm ci
      - name: Run tests
        working-directory: ${{ inputs.service-path }}
        run: npm test

  npm-audit:
    name: npm audit
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version-file: ${{ inputs.service-path }}/package.json
          cache: npm
          cache-dependency-path: ${{ inputs.service-path }}/package-lock.json
      - name: Install dependencies
        working-directory: ${{ inputs.service-path }}
        run: npm ci
      - name: Audit
        working-directory: ${{ inputs.service-path }}
        run: npm audit --audit-level=moderate

  build-image:
    name: Build image
    runs-on: ubuntu-latest
    needs: [lint, test, npm-audit]
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        if: inputs.push-image
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Extract image metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ inputs.image-name }}
          tags: |
            type=sha
            type=edge,branch=main
      - uses: docker/build-push-action@v6
        with:
          context: ${{ inputs.service-path }}
          push: ${{ inputs.push-image }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  secret-scan:
    name: Secret scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: trufflesecurity/trufflehog@v3
        with:
          extra_args: --only-verified

  trivy:
    name: Trivy scan
    runs-on: ubuntu-latest
    needs: build-image
    permissions:
      security-events: write
    steps:
      - uses: actions/checkout@v4
      - name: Scan with Trivy
        uses: aquasecurity/trivy-action@0.28.0
        with:
          scan-type: fs
          scan-ref: ${{ inputs.service-path }}
          format: sarif
          output: trivy-results.sarif
          severity: ${{ inputs.trivy-severity != '' && inputs.trivy-severity || 'CRITICAL,HIGH,MEDIUM,LOW,UNKNOWN' }}
          exit-code: ${{ inputs.trivy-severity != '' && '1' || '0' }}
      - name: Upload SARIF
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-results.sarif
```

- [ ] **Step 2: Validate YAML syntax**

```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/_node-service.yml'))" && echo "Valid YAML"
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/_node-service.yml
git commit -m "ci: add reusable Node service workflow"
```

---

## Task 5: CI Dispatcher Workflow

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Write ci.yml**

`.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: ['**']
  pull_request:
    branches: [main]

jobs:
  detect-changes:
    name: Detect changed paths
    runs-on: ubuntu-latest
    outputs:
      api: ${{ steps.filter.outputs.api }}
      web: ${{ steps.filter.outputs.web }}
    steps:
      - uses: actions/checkout@v4
      - uses: dorny/paths-filter@v3
        id: filter
        with:
          filters: |
            api:
              - 'services/api/**'
              - '.github/workflows/ci.yml'
              - '.github/workflows/_go-service.yml'
            web:
              - 'services/web/**'
              - '.github/workflows/ci.yml'
              - '.github/workflows/_node-service.yml'

  api-ci:
    name: API
    needs: detect-changes
    if: needs.detect-changes.outputs.api == 'true'
    uses: ./.github/workflows/_go-service.yml
    permissions:
      contents: read
      packages: write
      security-events: write
    with:
      service-path: services/api
      image-name: ghcr.io/lvermeire/api
      push-image: ${{ github.ref == 'refs/heads/main' && github.event_name == 'push' }}
      trivy-severity: ${{ github.ref == 'refs/heads/main' && github.event_name == 'push' && 'HIGH,CRITICAL' || '' }}

  web-ci:
    name: Web
    needs: detect-changes
    if: needs.detect-changes.outputs.web == 'true'
    uses: ./.github/workflows/_node-service.yml
    permissions:
      contents: read
      packages: write
      security-events: write
    with:
      service-path: services/web
      image-name: ghcr.io/lvermeire/web
      push-image: ${{ github.ref == 'refs/heads/main' && github.event_name == 'push' }}
      trivy-severity: ${{ github.ref == 'refs/heads/main' && github.event_name == 'push' && 'HIGH,CRITICAL' || '' }}
```

- [ ] **Step 2: Validate YAML syntax**

```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))" && echo "Valid YAML"
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add dispatcher workflow with path-based filtering"
```

---

## Task 6: Release Workflow

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Write release.yml**

`.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    branches: [main]
    tags: ['v*']

jobs:
  # Runs on every push to main.
  # Creates or updates a release PR (version bump + CHANGELOG).
  # When the release PR is merged, this job creates the git tag.
  release-please:
    name: Release please
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    permissions:
      contents: write
      pull-requests: write
    steps:
      - uses: google-github-actions/release-please-action@v4
        with:
          token: ${{ secrets.GITHUB_TOKEN }}

  # Runs when a release tag (v*) is pushed — i.e. after the release PR is merged.
  # Builds and pushes stable semver-tagged images.
  # Trivy blocks on MEDIUM+ before push.
  publish-api:
    name: Publish API image
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')
    permissions:
      contents: read
      packages: write
      security-events: write
    steps:
      - uses: actions/checkout@v4
      - name: Scan with Trivy (MEDIUM+ blocks)
        uses: aquasecurity/trivy-action@0.28.0
        with:
          scan-type: fs
          scan-ref: services/api
          format: sarif
          output: trivy-api.sarif
          severity: CRITICAL,HIGH,MEDIUM
          exit-code: '1'
      - name: Upload SARIF
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-api.sarif
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Extract image metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ghcr.io/lvermeire/api
          tags: |
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=semver,pattern={{major}}
      - uses: docker/build-push-action@v6
        with:
          context: services/api
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  publish-web:
    name: Publish Web image
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')
    permissions:
      contents: read
      packages: write
      security-events: write
    steps:
      - uses: actions/checkout@v4
      - name: Scan with Trivy (MEDIUM+ blocks)
        uses: aquasecurity/trivy-action@0.28.0
        with:
          scan-type: fs
          scan-ref: services/web
          format: sarif
          output: trivy-web.sarif
          severity: CRITICAL,HIGH,MEDIUM
          exit-code: '1'
      - name: Upload SARIF
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-web.sarif
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Extract image metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ghcr.io/lvermeire/web
          tags: |
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=semver,pattern={{major}}
      - uses: docker/build-push-action@v6
        with:
          context: services/web
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

- [ ] **Step 2: Validate YAML syntax**

```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))" && echo "Valid YAML"
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci: add release-please workflow with semver image publish"
```

---

## Task 7: Scheduled Security Workflow

**Files:**
- Create: `.github/workflows/security.yml`

- [ ] **Step 1: Write security.yml**

`.github/workflows/security.yml`:

```yaml
name: Security

on:
  schedule:
    - cron: '0 8 * * 1'  # Every Monday at 08:00 UTC
  workflow_dispatch:

jobs:
  trivy-api:
    name: Trivy — API
    runs-on: ubuntu-latest
    permissions:
      security-events: write
    steps:
      - uses: actions/checkout@v4
      - uses: aquasecurity/trivy-action@0.28.0
        with:
          scan-type: fs
          scan-ref: services/api
          format: sarif
          output: trivy-api.sarif
          severity: CRITICAL,HIGH,MEDIUM,LOW,UNKNOWN
          exit-code: '0'
      - uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: trivy-api.sarif

  trivy-web:
    name: Trivy — Web
    runs-on: ubuntu-latest
    permissions:
      security-events: write
    steps:
      - uses: actions/checkout@v4
      - uses: aquasecurity/trivy-action@0.28.0
        with:
          scan-type: fs
          scan-ref: services/web
          format: sarif
          output: trivy-web.sarif
          severity: CRITICAL,HIGH,MEDIUM,LOW,UNKNOWN
          exit-code: '0'
      - uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: trivy-web.sarif

  govulncheck:
    name: govulncheck
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: services/api/go.mod
          cache: true
      - uses: golang/govulncheck-action@v1
        with:
          work-dir: services/api

  npm-audit:
    name: npm audit
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version-file: services/web/package.json
          cache: npm
          cache-dependency-path: services/web/package-lock.json
      - name: Install dependencies
        working-directory: services/web
        run: npm ci
      - name: Audit
        working-directory: services/web
        run: npm audit --audit-level=moderate

  secret-scan:
    name: Secret scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: trufflesecurity/trufflehog@v3
        with:
          extra_args: --only-verified
```

- [ ] **Step 2: Validate YAML syntax**

```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/security.yml'))" && echo "Valid YAML"
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/security.yml
git commit -m "ci: add weekly scheduled security scan workflow"
```

---

## Task 8: Branch Protection Bootstrap Script

**Files:**
- Create: `scripts/bootstrap-github.sh`

- [ ] **Step 1: Write bootstrap-github.sh**

`scripts/bootstrap-github.sh`:

```bash
#!/usr/bin/env bash
# Configure branch protection for main via GitHub API.
# Run once after the first CI workflow execution so GitHub knows the check names.
#
# Usage:
#   export GH_REPO=lvermeire/dx-connect-ci-scaffold   # or set GITHUB_REPOSITORY
#   bash scripts/bootstrap-github.sh
#
# Required: gh CLI authenticated (gh auth login)

set -euo pipefail

REPO="${GH_REPO:-${GITHUB_REPOSITORY:-lvermeire/dx-connect-ci-scaffold}}"

echo "Configuring branch protection for main on ${REPO}..."

# Status check names are the job *display names* as they appear in GitHub UI.
# With reusable workflows, they are prefixed by the calling job name:
#   "API / Lint", "API / Test", "Web / Lint", etc.
# Run CI once first, then verify names at:
#   https://github.com/lvermeire/dx-connect-ci-scaffold/settings/branches
CHECKS=(
  "API / Lint"
  "API / Test"
  "API / govulncheck"
  "API / Secret scan"
  "API / Trivy scan"
  "Web / Lint"
  "Web / Test"
  "Web / npm audit"
  "Web / Secret scan"
  "Web / Trivy scan"
)

# Build the contexts JSON array
CONTEXTS_JSON=$(printf '%s\n' "${CHECKS[@]}" | jq -R . | jq -sc .)

gh api "repos/${REPO}/branches/main/protection" \
  --method PUT \
  --header "Accept: application/vnd.github+json" \
  --field "required_status_checks[strict]=true" \
  --field "required_status_checks[contexts]=${CONTEXTS_JSON}" \
  --field "enforce_admins=false" \
  --field "required_pull_request_reviews[required_approving_review_count]=1" \
  --field "required_pull_request_reviews[dismiss_stale_reviews]=true" \
  --field "restrictions=null" \
  --field "allow_force_pushes=false" \
  --field "allow_deletions=false"

echo "Branch protection configured for main."
```

- [ ] **Step 2: Make executable**

```bash
chmod +x scripts/bootstrap-github.sh
```

- [ ] **Step 3: Validate the script parses correctly**

```bash
bash -n scripts/bootstrap-github.sh && echo "Syntax OK"
```

- [ ] **Step 4: Commit**

```bash
git add scripts/bootstrap-github.sh
git commit -m "chore: add branch protection bootstrap script"
```

---

## Done

At this point the repo has:
- Reusable CI workflows for Go and Node services with lint, test, vuln scan, image build/push, secret scan, trivy
- Path-based dispatcher that only runs affected service CI
- Release-please managing the release PR and version bumping
- Semver image publishing on release tags with MEDIUM+ trivy gate
- Weekly scheduled security scans reporting to GitHub Security tab
- Renovate config for automated dependency updates with digest pinning
- Branch protection script documenting and applying required status checks

**Graduated trivy severity:**
- Feature branch / PR → informational (all severities shown, no block)
- Push to main → HIGH,CRITICAL blocks
- Release tag → MEDIUM+ blocks

**Next:** merge this PR to main, run `scripts/bootstrap-github.sh` to apply branch protection, then enable Renovate from the GitHub Marketplace.
