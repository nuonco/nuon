#!/bin/bash

set -euo pipefail

RUNNER_ID="${1:?Usage: $0 <runner_id> [--with-binary]}"
WITH_BINARY=false
for arg in "${@:2}"; do
  case "$arg" in
    --with-binary) WITH_BINARY=true ;;
  esac
done

CTL_API_URL="${CTL_API_URL:-http://localhost:8082}"
TTL="${TTL:-2h}"
POLL_TIMEOUT="${POLL_TIMEOUT:-600}"
POLL_INTERVAL="${POLL_INTERVAL:-15}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
MONO_ROOT="${MONO_ROOT:-$(cd "${REPO_ROOT}/../mono" 2>/dev/null && pwd || echo "")}"
TAG="$(uuidgen | tr '[:upper:]' '[:lower:]')"
IMAGE="ttl.sh/${TAG}:${TTL}"

AUTH_HEADER=""
if [ -n "${ADMIN_TOKEN:-}" ]; then
  AUTH_HEADER="Authorization: Bearer ${ADMIN_TOKEN}"
fi

admin_curl() {
  curl -sf ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$@"
}

cleanup() {
  if [ -n "${SERVE_DIR:-}" ] && [ -d "${SERVE_DIR:-}" ]; then
    rm -rf "$SERVE_DIR"
  fi
}
trap cleanup EXIT

echo "==> Building runner image via nctl..."
if [ -z "$MONO_ROOT" ] || [ ! -d "$MONO_ROOT" ]; then
  echo "❌ Cannot find mono repo. Set MONO_ROOT or place it at ../mono"
  exit 1
fi

(
  cd "$MONO_ROOT"
  RUN_NUONCTL_VERSION=1 RUN_NUONCTL_PATH=/dev/null \
    go run ./bins/nuonctl builds target runner --target final
)

echo "==> Tagging and pushing image to ttl.sh (expires in ${TTL})..."
docker tag dev.nuon.co/runner:final "${IMAGE}"
docker push "${IMAGE}"

SETTINGS="{\"container_image_url\": \"ttl.sh/${TAG}\", \"container_image_tag\": \"${TTL}\"}"

if [ "$WITH_BINARY" = true ]; then
  SERVE_DIR="$(mktemp -d)"
  BINARY_NAME="runner_linux_amd64"

  echo "==> Cross-compiling runner binary for linux/amd64..."
  (cd "$REPO_ROOT" && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "${SERVE_DIR}/${BINARY_NAME}" ./bins/runner)

  AZURE_SA="${AZURE_SA:-}"
  AZURE_RG="${AZURE_RG:-}"
  CONTAINER="runner"

  if [ -z "$AZURE_RG" ]; then
    echo "==> Auto-detecting Azure resource group..."
    AZURE_RG=$(az vmss list --query "[0].resourceGroup" -o tsv 2>/dev/null || true)
    if [ -z "$AZURE_RG" ]; then
      echo "❌ Could not detect Azure resource group. Set AZURE_RG."
      exit 1
    fi
  fi

  if [ -z "$AZURE_SA" ]; then
    AZURE_SA=$(az storage account list --resource-group "$AZURE_RG" \
      --query "[?starts_with(name, 'nuondevrunner')].name | [0]" -o tsv 2>/dev/null || true)

    if [ -z "$AZURE_SA" ] || [ "$AZURE_SA" = "null" ]; then
      AZURE_SA="nuondevrunner$(date +%s | tail -c 6)"
      LOCATION=$(az group show --name "$AZURE_RG" --query location -o tsv 2>/dev/null)
      echo "==> Creating storage account ${AZURE_SA} in ${AZURE_RG}..."
      az storage account create \
        --name "$AZURE_SA" \
        --resource-group "$AZURE_RG" \
        --location "$LOCATION" \
        --sku Standard_LRS \
        --min-tls-version TLS1_2 \
        -o none 2>/dev/null
    fi
  fi

  az storage container create \
    --name "$CONTAINER" \
    --account-name "$AZURE_SA" \
    --auth-mode key \
    -o none 2>/dev/null || true

  echo "==> Uploading binary to Azure Blob Storage (${AZURE_SA})..."
  az storage blob upload \
    --account-name "$AZURE_SA" \
    --container-name "$CONTAINER" \
    --name "$BINARY_NAME" \
    --file "${SERVE_DIR}/${BINARY_NAME}" \
    --auth-mode key \
    --overwrite \
    -o none 2>/dev/null

  EXPIRY=$(date -u -v+2H +%Y-%m-%dT%H:%MZ 2>/dev/null || date -u -d '+2 hours' +%Y-%m-%dT%H:%MZ)
  BINARY_URL=$(az storage blob generate-sas \
    --account-name "$AZURE_SA" \
    --container-name "$CONTAINER" \
    --name "$BINARY_NAME" \
    --permissions r \
    --expiry "$EXPIRY" \
    --auth-mode key \
    --full-uri \
    -o tsv 2>/dev/null)

  echo "==> Binary uploaded: ${AZURE_SA}/${CONTAINER}/${BINARY_NAME}"

  SETTINGS="{\"container_image_url\": \"ttl.sh/${TAG}\", \"container_image_tag\": \"${TTL}\", \"runner_binary_url\": \"${BINARY_URL}\"}"
fi

echo "==> Updating runner settings..."
admin_curl -X PATCH "${CTL_API_URL}/v1/runners/${RUNNER_ID}/settings" \
  -H "Content-Type: application/json" \
  -d "${SETTINGS}" \
  | jq .

echo ""
echo "✅ Artifacts ready!"
echo "   image:  ${IMAGE}"
if [ "$WITH_BINARY" = true ]; then
  echo "   binary: ${AZURE_SA}/${CONTAINER}/${BINARY_NAME} (SAS URL expires in 2h)"
fi
echo "   runner: ${RUNNER_ID}"
echo ""

if [ "$WITH_BINARY" = true ]; then
  AZURE_VMSS="${AZURE_VMSS:-}"

  if [ -z "$AZURE_VMSS" ]; then
    echo "==> Auto-detecting VMSS in ${AZURE_RG}..."
    AZURE_VMSS=$(az vmss list --resource-group "$AZURE_RG" \
      --query "[0].name" -o tsv 2>/dev/null || true)
    if [ -z "$AZURE_VMSS" ] || [ "$AZURE_VMSS" = "null" ]; then
      echo "❌ Could not detect VMSS. Set AZURE_VMSS."
      exit 1
    fi
  fi

  INSTANCE_IDS="${AZURE_VMSS_IDS:-}"
  if [ -z "$INSTANCE_IDS" ]; then
    INSTANCE_IDS=$(az vmss list-instances --resource-group "$AZURE_RG" \
      --name "$AZURE_VMSS" --query "[].instanceId" -o tsv 2>/dev/null || true)
    if [ -z "$INSTANCE_IDS" ]; then
      echo "❌ No VMSS instances found in ${AZURE_VMSS}."
      exit 1
    fi
  fi

  for INSTANCE_ID in $INSTANCE_IDS; do
    echo "==> Deploying binary to ${AZURE_VMSS} instance ${INSTANCE_ID}..."
    DEPLOY_OUTPUT=$(az vmss run-command invoke \
      --resource-group "$AZURE_RG" \
      --name "$AZURE_VMSS" \
      --instance-id "$INSTANCE_ID" \
      --command-id RunShellScript \
      --scripts "
        set -e
        curl -sSL '${BINARY_URL}' -o /tmp/runner_new
        chmod +x /tmp/runner_new
        mv /tmp/runner_new /opt/nuon/runner/bin/runner
        systemctl restart nuon-runner-mng.service
        echo 'runner binary updated and mng service restarted'
      " 2>&1) || true

    echo "$DEPLOY_OUTPUT" | jq -r '.value[0].message // empty' 2>/dev/null || echo "$DEPLOY_OUTPUT"
  done

  echo ""
  echo "The script will wait for the runner to come online..."
else
  echo "Provision or reprovision the install now."
  echo "The script will wait for the runner to come online..."
fi
echo ""

ELAPSED=0
LAST_STATUS=""

while [ "$ELAPSED" -lt "$POLL_TIMEOUT" ]; do
  STATUS=$(admin_curl "${CTL_API_URL}/v1/runners/${RUNNER_ID}" 2>/dev/null \
    | jq -r '.status // empty' 2>/dev/null || true)

  if [ "$STATUS" != "$LAST_STATUS" ] && [ -n "$STATUS" ]; then
    echo "   runner status: ${STATUS} (${ELAPSED}s)"
    LAST_STATUS="$STATUS"
  fi

  if [ "$STATUS" = "active" ]; then
    echo ""
    echo "✅ Runner is active!"

    CURRENT_TAG=$(admin_curl "${CTL_API_URL}/v1/runners/${RUNNER_ID}/settings" 2>/dev/null \
      | jq -r '.container_image_tag // empty' 2>/dev/null || true)

    if [ "$CURRENT_TAG" = "$TTL" ]; then
      echo "   confirmed: container_image_tag=${CURRENT_TAG}"
    else
      echo "   ⚠️  container_image_tag is '${CURRENT_TAG}' (expected '${TTL}')"
    fi
    break
  fi

  sleep "$POLL_INTERVAL"
  ELAPSED=$((ELAPSED + POLL_INTERVAL))
done

if [ "$ELAPSED" -ge "$POLL_TIMEOUT" ]; then
  echo ""
  echo "⏰ Timed out after ${POLL_TIMEOUT}s. Last status: ${LAST_STATUS:-unknown}"
fi

echo ""
echo "🧹 Cleaning up..."
