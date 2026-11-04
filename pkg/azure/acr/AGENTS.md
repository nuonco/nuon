# ACR

Entra tokens minted here use `RegistryScope` (`https://containerregistry.azure.net/.default`).

`GetRepositoryTokenWithCredential` requests that scope. The resulting access token is posted to the
registry login server. Callers that verify an ACR cloud connection, including the foreign-subject
probe, request the same scope.

Do not request an Azure Resource Manager management scope (`https://management.azure.com/` or
`https://management.azure.com/.default`) for ACR pull, exchange, or verification.
