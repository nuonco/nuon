resource "azurerm_resource_group" "main" {
  name     = var.env
  location = "East US"
}

resource "azurerm_policy_definition" "ensure-location" {
  name         = "ensure-location"
  policy_type  = "Custom"
  mode         = "All"
  display_name = "my-policy-definition"

  policy_rule = <<POLICY_RULE
 {
    "if": {
      "not": {
        "field": "location",
        "equals": "eastus"
      }
    },
    "then": {
      "effect": "Deny"
    }
  }
POLICY_RULE
}

resource "azurerm_resource_group_policy_assignment" "main" {
  name                 = "main"
  resource_group_id    = azurerm_resource_group.main.id
  policy_definition_id = azurerm_policy_definition.ensure-location.id

  parameters = <<PARAMS
    {
      "tagName": {
        "value": "Env"
      },
      "tagValue": {
        "value": "${var.env}"
      }
    }
PARAMS
}
