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

# Only the gate job is required. Individual service jobs (API / Lint, Web / Test,
# etc.) are visible in PRs but not blocking — they are skipped when their paths
# didn't change, which would otherwise prevent unrelated PRs from merging.
CHECKS=("CI")

# Build the full JSON body — gh api --field cannot send nested arrays correctly
CONTEXTS_JSON=$(printf '%s\n' "${CHECKS[@]}" | jq -R . | jq -sc .)

BODY=$(jq -n \
  --argjson contexts "$CONTEXTS_JSON" \
  '{
    required_status_checks: { strict: true, contexts: $contexts },
    enforce_admins: false,
    required_pull_request_reviews: {
      required_approving_review_count: 1,
      dismiss_stale_reviews: true
    },
    restrictions: null,
    allow_force_pushes: false,
    allow_deletions: false
  }')

echo "$BODY" | gh api "repos/${REPO}/branches/main/protection" \
  --method PUT \
  --header "Accept: application/vnd.github+json" \
  --input -

echo "Branch protection configured for main."
