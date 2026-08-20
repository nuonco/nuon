package arm

import "fmt"

// getPhoneHomeIdentityResource creates the identity the phone-home script authenticates
// as. Deliberately carries no role assignments: it never touches Azure, so a token minted
// for it is inert everywhere except the phone-home endpoint.
func getPhoneHomeIdentityResource(identityName string, scope armScope) map[string]any {
	return map[string]any{
		"type":       "Microsoft.ManagedIdentity/userAssignedIdentities",
		"apiVersion": "2023-01-31",
		"name":       identityName,
		"location":   "[parameters('location')]",
		"tags":       scope.innerCommonTagsExpr(),
	}
}

func phoneHomeIdentityResourceID(identityName string) string {
	return fmt.Sprintf(
		"[resourceId('Microsoft.ManagedIdentity/userAssignedIdentities', '%s')]", identityName)
}

func phoneHomeIdentityClientID(identityName string) string {
	return fmt.Sprintf(
		"[reference(resourceId('Microsoft.ManagedIdentity/userAssignedIdentities', '%s'), '2023-01-31').clientId]",
		identityName)
}

// phoneHomeAuthScript fetches a token straight from IMDS rather than going through
// `az login --identity`, which enumerates subscriptions and fails for an identity holding
// no role assignments.
//
// The token is written to a curl config file instead of an argv header so it stays out of
// the process list, and nothing echoes it: deployment script logs are readable by any
// reader on the resource group.
const phoneHomeAuthScript = `
IMDS_URL="http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fmanagement.azure.com%2F&client_id=${PHONE_HOME_IDENTITY_CLIENT_ID}"

CURL_CONFIG=$(mktemp)
chmod 600 "$CURL_CONFIG"
trap 'rm -f "$CURL_CONFIG"' EXIT

# A freshly created identity is not immediately usable, so this retries rather than
# failing the deployment on a propagation delay.
TOKEN=""
for attempt in $(seq 1 12); do
  # Whitespace around the colon is tolerated: the IMDS response is JSON, so its exact
  # spacing is not part of the contract.
  TOKEN=$(curl -s -f -H "Metadata: true" "$IMDS_URL" |
    sed -n 's/.*"access_token"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
  if [ -n "$TOKEN" ]; then
    break
  fi
  echo "waiting for managed identity token (attempt $attempt)"
  sleep 10
done

if [ -z "$TOKEN" ]; then
  echo "failed to acquire managed identity token"
  exit 1
fi

printf 'header = "Authorization: Bearer %s"\n' "$TOKEN" > "$CURL_CONFIG"
unset TOKEN
`
