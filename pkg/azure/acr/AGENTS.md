# ACR

Cloud-connection tokens use `RegistryScope` (`https://containerregistry.azure.net/.default`).

`GetRepositoryTokenWithCredential` requests that scope and posts the access token with
`ConnectionExchangeClient`. Callers that verify an ACR cloud connection, including the
foreign-subject probe, request the same scope. Do not request an Azure Resource Manager management
scope (`https://management.azure.com/` or `https://management.azure.com/.default`) for those calls.

Legacy `GetRepositoryToken` keeps `armTokenScope` (`https://management.azure.com/.default`) and
`exchangeClient`, the default HTTP client. Existing runners behind a proxy depend on that. Do not
move the legacy path onto `RegistryScope` or `ConnectionExchangeClient`.
