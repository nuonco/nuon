#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# Local development script for ctl-api
# Starts docker-compose dependencies and runs ctl-api with local config
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# ---------------------------------------------------------------------------
# Ensure infrastructure is running (via ~/nuon/mono docker-compose)
# ---------------------------------------------------------------------------
MONO_ROOT="${MONO_ROOT:-$HOME/nuon/mono}"
if [ ! -f "$MONO_ROOT/docker-compose.yml" ]; then
  echo "ERROR: Could not find docker-compose.yml at $MONO_ROOT"
  echo "Set MONO_ROOT to point to your nuon/mono checkout."
  exit 1
fi

echo "Starting local infrastructure from $MONO_ROOT..."
docker compose -f "$MONO_ROOT/docker-compose.yml" up -d postgres temporal seaweedfs clickhouse-01 clickhouse-keeper-01 clickhouse-keeper-02 clickhouse-keeper-03

echo "Waiting for services to be ready..."

# Wait for PostgreSQL
until docker compose -f "$MONO_ROOT/docker-compose.yml" exec -T postgres pg_isready -U ctl_api 2>/dev/null; do
  sleep 1
done
echo "PostgreSQL is ready"

# Wait for SeaweedFS S3
for i in $(seq 1 30); do
  if curl -sf http://localhost:8333/ > /dev/null 2>&1; then
    break
  fi
  sleep 1
done
echo "SeaweedFS S3 is ready"

# ---------------------------------------------------------------------------
# Environment variables for ctl-api
# ---------------------------------------------------------------------------

# --- Core service identity ---
export ENV=development
export SERVICE_NAME=ctl-api
export SERVICE_TYPE=api
export SERVICE_DEPLOYMENT=local
export GIT_REF=local
export VERSION=local

# --- HTTP ports ---
export HTTP_PORT=8181
export INTERNAL_HTTP_PORT=8182
export RUNNER_HTTP_PORT=8183
export AUTH_HTTP_PORT=8184
export ADMIN_DASHBOARD_HTTP_PORT=8185
export GRACEFUL_SHUTDOWN_TIMEOUT=5s

# --- PostgreSQL ---
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=ctl_api
export DB_PASSWORD=ctl_api
export DB_NAME=ctl_api
export DB_SSL_MODE=disable
export DB_REGION=us-west-2

# --- ClickHouse ---
export CLICKHOUSE_DB_HOST=localhost
export CLICKHOUSE_DB_PORT=9000
export CLICKHOUSE_DB_USER=ctl_api
export CLICKHOUSE_DB_PASSWORD=ctl_api
export CLICKHOUSE_DB_NAME=ctl_api

# --- Temporal ---
export TEMPORAL_HOST=localhost:7233

# --- Local mode (points S3 at SeaweedFS, bypasses real AWS) ---
export USE_LOCAL=true
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=local
export AWS_SECRET_ACCESS_KEY=local

# --- CloudFormation S3 bucket ---
export AWS_CLOUDFORMATION_STACK_TEMPLATE_BUCKET=cfn-templates-local
export AWS_CLOUDFORMATION_STACK_TEMPLATE_BUCKET_REGION=us-east-1
export AWS_CLOUDFORMATION_STACK_TEMPLATE_BASE_URL=http://localhost:8333/cfn-templates-local

# --- Auth0 (placeholder values for local dev) ---
export AUTH0_ISSUER_URL=https://localhost/
export AUTH0_AUDIENCE=local-audience
export AUTH0_CLIENT_ID=local-client-id

# --- GitHub App (placeholder values) ---
export GITHUB_APP_ID=000000
export GITHUB_APP_KEY=placeholder-key

# --- URLs ---
export APP_URL=http://localhost:4000
export RUNNER_API_URL=http://localhost:8083
export PUBLIC_API_URL=http://localhost:8081
export ADMIN_API_URL=http://localhost:8082
export TEMPORAL_UI_URL=http://localhost:8233

# --- Sandbox / Local mode ---
export FORCE_SANDBOX_MODE=true
export SANDBOX_MODE_ENABLE_RUNNERS=false

# --- Notifications (disabled) ---
export LOOPS_API_KEY=placeholder-loops-key
export INTERNAL_SLACK_WEBHOOK_URL=https://hooks.slack.com/placeholder
export DISABLE_NOTIFICATIONS=true

# --- Runner config (placeholders) ---
export RUNNER_CONTAINER_IMAGE_TAG=latest
export RUNNER_DEFAULT_SUPPORT_IAM_ROLE_ARN=arn:aws:iam::000000000000:role/placeholder
export USE_LOCAL_RUNNERS=true

# --- Management AWS config (placeholders) ---
export MANAGEMENT_IAM_ROLE_ARN=arn:aws:iam::000000000000:role/placeholder
export MANAGEMENT_ACCOUNT_ID=000000000000
export MANAGEMENT_ECR_REGISTRY_ID=000000000000
export MANAGEMENT_ECR_REGISTRY_ARN=arn:aws:ecr:us-west-2:000000000000:registry

# --- Org Runner config (placeholders) ---
export ORG_RUNNER_K8S_CLUSTER_ID=placeholder-cluster
export ORG_RUNNER_K8S_PUBLIC_ENDPOINT=https://localhost:6443
export ORG_RUNNER_K8S_CA_DATA=placeholder-ca-data
export ORG_RUNNER_OIDC_PROVIDER_URL=https://localhost
export ORG_RUNNER_OIDC_PROVIDER_ARN=arn:aws:iam::000000000000:oidc-provider/placeholder
export ORG_RUNNER_REGION=us-west-2
export ORG_RUNNER_SUPPORT_ROLE_ARN=arn:aws:iam::000000000000:role/placeholder
export ORG_RUNNER_K8S_IAM_ROLE_ARN=arn:aws:iam::000000000000:role/placeholder

# --- DNS config (placeholders) ---
export DNS_MANAGEMENT_IAM_ROLE_ARN=arn:aws:iam::000000000000:role/placeholder
export DNS_ZONE_ID=Z00000000000000000000
export DNS_ROOT_DOMAIN=local.nuon.dev

# --- Analytics (disabled) ---
export SEGMENT_WRITE_KEY=placeholder-segment-key
export DISABLE_ANALYTICS=true

# --- Integration (placeholder) ---
export INTEGRATION_GITHUB_INSTALL_ID=000000

# --- Metrics (disabled) ---
export DISABLE_METRICS=true

# ---------------------------------------------------------------------------
# Run startup (migrations) then API
# ---------------------------------------------------------------------------
echo ""
echo "Running database migrations..."
go run "$REPO_ROOT/services/ctl-api/..." startup

echo ""
echo "Starting ctl-api (all APIs)..."
go run "$REPO_ROOT/services/ctl-api/..." api
