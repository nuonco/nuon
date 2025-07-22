resource "azurerm_storage_account" "main" {
  name                     = var.name
  resource_group_name      = data.azurerm_resource_group.current.name
  location                 = data.azurerm_resource_group.current.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

