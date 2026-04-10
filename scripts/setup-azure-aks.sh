#!/bin/bash
#
# Set up an Azure AKS cluster for local ctl-api development.
#
# This script creates (or reuses) an AKS cluster with the features required
# by ctl-api's org-runner provisioning flow:
#   - OIDC Issuer + Workload Identity (for runner pod auth)
#   - Azure RBAC enabled
#   - kubelogin support for local kubectl / ctl-api access
#
# It outputs the environment variables ctl-api needs to connect to the cluster
# (OrgRunnerK8s* fields from config.go).
#
# Prerequisites:
#   - Azure CLI (`az`) installed and logged in
#   - kubelogin installed (brew install Azure/kubelogin/kubelogin)
#   - A resource group (can be the same one used by setup-azure-acr.sh)
#
# Usage:
#   ./scripts/setup-azure-aks.sh                           # interactive prompts
#   ./scripts/setup-azure-aks.sh --rg mygroup              # specify resource group
#   ./scripts/setup-azure-aks.sh --rg mygroup --name myaks --location eastus
#
# After running, source the generated env file:
#   source /tmp/nuon-azure-aks.env
#
# Combine with ACR setup:
#   source /tmp/nuon-azure-acr.env
#   source /tmp/nuon-azure-aks.env
#   go run services/ctl-api/main.go api
#

set -euo pipefail

# ── parse args ────────────────────────────────────────────────────────
RESOURCE_GROUP=""
CLUSTER_NAME=""
LOCATION=""
NODE_COUNT="1"
NODE_VM_SIZE="Standard_D2s_v3"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --rg)          RESOURCE_GROUP="$2"; shift 2 ;;
    --name)        CLUSTER_NAME="$2"; shift 2 ;;
    --location)    LOCATION="$2"; shift 2 ;;
    --node-count)  NODE_COUNT="$2"; shift 2 ;;
    --vm-size)     NODE_VM_SIZE="$2"; shift 2 ;;
    -h|--help)
      echo "Usage: $0 [--rg RESOURCE_GROUP] [--name CLUSTER_NAME] [--location LOCATION] [--node-count N] [--vm-size SIZE]"
      exit 0
      ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# ── check prerequisites ──────────────────────────────────────────────
if ! command -v az &>/dev/null; then
  echo "❌ Azure CLI (az) is not installed. Install it first:"
  echo "   brew install azure-cli"
  exit 1
fi

if ! command -v kubelogin &>/dev/null; then
  echo "❌ kubelogin is not installed. Install it first:"
  echo "   brew install Azure/kubelogin/kubelogin"
  exit 1
fi

if ! az account show &>/dev/null; then
  echo "❌ Not logged in to Azure. Run: az login"
  exit 1
fi

ACCOUNT_INFO=$(az account show -o json)
SUBSCRIPTION_ID=$(echo "$ACCOUNT_INFO" | jq -r '.id')
TENANT_ID=$(echo "$ACCOUNT_INFO" | jq -r '.tenantId')
ACCOUNT_NAME=$(echo "$ACCOUNT_INFO" | jq -r '.name')

echo "Azure account: ${ACCOUNT_NAME}"
echo "Subscription:  ${SUBSCRIPTION_ID}"
echo "Tenant:        ${TENANT_ID}"
echo ""

# ── resolve resource group ───────────────────────────────────────────
if [ -z "$RESOURCE_GROUP" ]; then
  echo "Available resource groups:"
  az group list --query "[].name" -o tsv | head -20
  echo ""
  read -rp "Resource group (or enter a new name to create): " RESOURCE_GROUP
fi

if ! az group show --name "$RESOURCE_GROUP" &>/dev/null; then
  if [ -z "$LOCATION" ]; then
    read -rp "Resource group doesn't exist. Location to create it in (e.g. eastus): " LOCATION
  fi
  echo "==> Creating resource group ${RESOURCE_GROUP} in ${LOCATION}..."
  az group create --name "$RESOURCE_GROUP" --location "$LOCATION" -o none
else
  if [ -z "$LOCATION" ]; then
    LOCATION=$(az group show --name "$RESOURCE_GROUP" --query location -o tsv)
  fi
fi

echo "Resource group: ${RESOURCE_GROUP} (${LOCATION})"

# ── resolve cluster name ─────────────────────────────────────────────
if [ -z "$CLUSTER_NAME" ]; then
  # Check for existing AKS cluster in the resource group
  EXISTING_CLUSTER=$(az aks list --resource-group "$RESOURCE_GROUP" \
    --query "[0].name" -o tsv 2>/dev/null || true)

  if [ -n "$EXISTING_CLUSTER" ] && [ "$EXISTING_CLUSTER" != "null" ]; then
    echo "Found existing AKS cluster: ${EXISTING_CLUSTER}"
    read -rp "Use it? [Y/n]: " USE_EXISTING
    if [[ "${USE_EXISTING:-Y}" =~ ^[Yy]?$ ]]; then
      CLUSTER_NAME="$EXISTING_CLUSTER"
    fi
  fi

  if [ -z "$CLUSTER_NAME" ]; then
    DEFAULT_NAME="nuon-dev-$(whoami | tr -cd '[:alnum:]' | head -c 6)"
    read -rp "AKS cluster name [${DEFAULT_NAME}]: " CLUSTER_NAME
    CLUSTER_NAME="${CLUSTER_NAME:-$DEFAULT_NAME}"
  fi
fi

# ── create or reuse AKS cluster ─────────────────────────────────────
if az aks show --name "$CLUSTER_NAME" --resource-group "$RESOURCE_GROUP" &>/dev/null; then
  echo "==> Using existing AKS cluster: ${CLUSTER_NAME}"

  # Check if required features are already enabled
  CLUSTER_FEATURES=$(az aks show --name "$CLUSTER_NAME" --resource-group "$RESOURCE_GROUP" \
    --query "{oidc:oidcIssuerProfile.enabled, wi:securityProfile.workloadIdentity.enabled, nap:nodeProvisioningProfile.mode}" -o json)
  HAS_OIDC=$(echo "$CLUSTER_FEATURES" | jq -r '.oidc // false')
  HAS_WI=$(echo "$CLUSTER_FEATURES" | jq -r '.wi // false')
  HAS_NAP=$(echo "$CLUSTER_FEATURES" | jq -r '.nap // "Manual"')

  if [ "$HAS_OIDC" = "true" ] && [ "$HAS_WI" = "true" ] && [ "$HAS_NAP" = "Auto" ]; then
    echo "==> OIDC issuer, workload identity, and NAP already enabled. Skipping update."
  else
    echo "==> Enabling OIDC issuer, workload identity, and NAP (this may take a few minutes)..."
    az aks update \
      --name "$CLUSTER_NAME" \
      --resource-group "$RESOURCE_GROUP" \
      --enable-oidc-issuer \
      --enable-workload-identity \
      --node-provisioning-mode Auto \
      -o none
  fi
else
  echo "==> Creating AKS cluster: ${CLUSTER_NAME}..."
  echo "   VM size:    ${NODE_VM_SIZE}"
  echo "   Node count: ${NODE_COUNT}"
  echo "   This may take 5-10 minutes..."

  az aks create \
    --name "$CLUSTER_NAME" \
    --resource-group "$RESOURCE_GROUP" \
    --location "$LOCATION" \
    --node-count "$NODE_COUNT" \
    --node-vm-size "$NODE_VM_SIZE" \
    --enable-oidc-issuer \
    --enable-workload-identity \
    --enable-azure-rbac \
    --enable-aad \
    --network-plugin azure \
    --network-plugin-mode overlay \
    --node-provisioning-mode Auto \
    --generate-ssh-keys \
    -o none

  echo "   AKS cluster created."
fi

# ── get cluster details ──────────────────────────────────────────────
echo "==> Fetching cluster details..."

CLUSTER_INFO=$(az aks show --name "$CLUSTER_NAME" --resource-group "$RESOURCE_GROUP" -o json)

CLUSTER_ID=$(echo "$CLUSTER_INFO" | jq -r '.id')
API_SERVER=$(echo "$CLUSTER_INFO" | jq -r '.fqdn')
API_ENDPOINT="https://${API_SERVER}:443"
OIDC_ISSUER_URL=$(echo "$CLUSTER_INFO" | jq -r '.oidcIssuerProfile.issuerUrl')

# Get the CA data from cluster credentials via temp kubeconfig
TEMP_KUBECONFIG=$(mktemp)
trap "rm -f $TEMP_KUBECONFIG" EXIT
az aks get-credentials \
  --name "$CLUSTER_NAME" \
  --resource-group "$RESOURCE_GROUP" \
  --file "$TEMP_KUBECONFIG" \
  --overwrite-existing \
  -o none

CA_DATA=$(kubectl config view --raw --kubeconfig "$TEMP_KUBECONFIG" \
  -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')
API_ENDPOINT=$(kubectl config view --raw --kubeconfig "$TEMP_KUBECONFIG" \
  -o jsonpath='{.clusters[0].cluster.server}')

echo ""
echo "Cluster name:  ${CLUSTER_NAME}"
echo "API endpoint:  ${API_ENDPOINT}"
echo "CA data:       ${CA_DATA:0:40}..."

# ── assign current user AKS RBAC Cluster Admin ───────────────────────
echo ""
echo "==> Ensuring current user has AKS RBAC Cluster Admin role..."

USER_OBJECT_ID=$(az ad signed-in-user show --query id -o tsv 2>/dev/null || true)
if [ -n "$USER_OBJECT_ID" ]; then
  az role assignment create \
    --assignee-object-id "$USER_OBJECT_ID" \
    --assignee-principal-type "User" \
    --role "Azure Kubernetes Service RBAC Cluster Admin" \
    --scope "$CLUSTER_ID" \
    -o none 2>/dev/null || true
  echo "   Role assigned (or already exists)."
else
  echo "   ⚠️  Could not determine current user object ID. You may need to assign"
  echo "   'Azure Kubernetes Service RBAC Cluster Admin' manually."
fi

# ── merge kubeconfig for kubectl access ──────────────────────────────
echo ""
echo "==> Merging cluster credentials into ~/.kube/config..."
az aks get-credentials \
  --name "$CLUSTER_NAME" \
  --resource-group "$RESOURCE_GROUP" \
  --overwrite-existing \
  -o none

# Convert kubeconfig to use kubelogin
kubelogin convert-kubeconfig -l azurecli 2>/dev/null || true
echo "   Done. You can now use kubectl against this cluster."

# ── attach ACR if one exists in the same resource group ──────────────
EXISTING_ACR=$(az acr list --resource-group "$RESOURCE_GROUP" \
  --query "[0].name" -o tsv 2>/dev/null || true)
if [ -n "$EXISTING_ACR" ] && [ "$EXISTING_ACR" != "null" ]; then
  echo ""
  echo "==> Attaching ACR ${EXISTING_ACR} to AKS cluster (for image pulls)..."
  az aks update \
    --name "$CLUSTER_NAME" \
    --resource-group "$RESOURCE_GROUP" \
    --attach-acr "$EXISTING_ACR" \
    -o none 2>/dev/null || echo "   (already attached or insufficient permissions)"
fi

# ── write env file ───────────────────────────────────────────────────
ENV_FILE="/tmp/nuon-azure-aks.env"

cat > "$ENV_FILE" << EOF
# Azure AKS org-runner config for ctl-api local development
# Generated by setup-azure-aks.sh on $(date -u +%Y-%m-%dT%H:%M:%SZ)
#
# Source this file before starting ctl-api:
#   source ${ENV_FILE}

# Org runner K8s cluster connection
export ORG_RUNNER_K8S_CLUSTER_ID=${CLUSTER_NAME}
export ORG_RUNNER_K8S_PUBLIC_ENDPOINT=${API_ENDPOINT}
export ORG_RUNNER_K8S_CA_DATA=${CA_DATA}
export ORG_RUNNER_REGION=${LOCATION}

# Azure identity management (used by org IAM provisioning)
export MANAGEMENT_AZURE_TENANT_ID=${TENANT_ID}
export MANAGEMENT_AZURE_SUBSCRIPTION_ID=${SUBSCRIPTION_ID}
export MANAGEMENT_AZURE_RESOURCE_GROUP=${RESOURCE_GROUP}
export MANAGEMENT_AZURE_OIDC_ISSUER_URL=${OIDC_ISSUER_URL}
EOF

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Azure AKS cluster is ready!"
echo ""
echo "Environment file written to: ${ENV_FILE}"
echo ""
cat "$ENV_FILE" | grep "^export"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "To use (combine with ACR setup):"
echo "  source /tmp/nuon-azure-acr.env"
echo "  source ${ENV_FILE}"
echo "  go run services/ctl-api/main.go api"
echo ""
echo "To verify cluster access:"
echo "  kubectl get nodes"
echo ""
