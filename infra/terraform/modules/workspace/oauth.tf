locals {
  # you can get this using the following api request
  # curl --header "Authorization: Bearer $TFE_TOKEN" \
  # --header "Content-Type: application/vnd.api+json" \
  # --request GET \
  # https://app.terraform.io/api/v2/organizations/nuonco/oauth-clients | jq .
  oauth_client_id = "oc-ZaFpTCTCdPdoDDtM"
}

data "tfe_oauth_client" "github" {
  oauth_client_id = local.oauth_client_id
}

