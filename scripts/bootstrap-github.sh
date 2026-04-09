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
