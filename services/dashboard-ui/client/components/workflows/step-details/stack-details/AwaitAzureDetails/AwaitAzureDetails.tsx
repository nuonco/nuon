import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Expand } from '@/components/common/Expand'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { TAppSecretConfig } from '@/types'
import type { IStackDetails } from '../types'

interface IAwaitAzureDetails extends IStackDetails {
  installId: string
  azureLocation?: string
  secrets?: TAppSecretConfig[]
}

export const AwaitAzureDetails = ({ stack, installId, azureLocation, secrets }: IAwaitAzureDetails) => {
  const vaultName = installId.slice(0, 24)
  const customerSecrets = secrets?.filter((s) => !s.auto_generate)
  const requiredSecrets = customerSecrets?.filter((s) => s.required || (!s.default && !s.required))
  const overridableSecrets = customerSecrets?.filter((s) => !s.required && !!s.default)

  const renderSecretCard = (secret: TAppSecretConfig) => {
    const kvName = secret.name.replaceAll('_', '-')
    const cmd = `az keyvault secret set --vault-name ${vaultName} --name ${kvName} --value "<your-secret-value>"`
    return (
      <Card key={secret.name}>
        <span className="flex justify-between items-center">
          <Text>
            {secret.display_name || secret.name}
            {secret.required && (
              <span className="text-red-500 ml-1">*</span>
            )}
          </Text>
          <ClickToCopyButton
            className="w-fit self-end"
            textToCopy={cmd}
          />
        </span>
        {secret.description && (
          <Text variant="subtext">{secret.description}</Text>
        )}
        <Code>{cmd}</Code>
      </Card>
    )
  }

  return (
    <>
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Provision the install stack using the Azure CLI
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Ensure you are logged into the Azure subscription you want to
              install into
            </Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={`az login`}
            />
          </span>
          <Code>az login</Code>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Create a resource group to deploy into</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={`az group create --name ${installId}-rg --location ${azureLocation}`}
            />
          </span>
          <Code>{`
            az group create --name ${installId}-rg --location ${azureLocation}
          `}</Code>
        </Card>
      </div>

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Create the Key Vault
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Create a Key Vault in the resource group</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={`az keyvault create --name ${vaultName} --resource-group ${installId}-rg --location ${azureLocation} --enable-rbac-authorization`}
            />
          </span>
          <Code>{`
            az keyvault create --name ${vaultName} --resource-group ${installId}-rg --location ${azureLocation} --enable-rbac-authorization
          `}</Code>
        </Card>
      </div>

      {customerSecrets && customerSecrets.length > 0 && (
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Create secrets in the Key Vault
          </Text>
          <Text variant="subtext">
            Before deploying the stack, create the following secrets in the Key
             Vault. The secret names must match exactly.
          </Text>

          <Card>
            <span className="flex justify-between items-center">
              <Text>
                Grant yourself permission to set secrets
              </Text>
              <ClickToCopyButton
                className="w-fit self-end"
                textToCopy={`az role assignment create --assignee "$(az ad signed-in-user show --query id -o tsv)" --role "Key Vault Secrets Officer" --scope "$(az keyvault show --name ${vaultName} --resource-group ${installId}-rg --query id -o tsv)"`}
              />
            </span>
            <Code>{`
              az role assignment create --assignee "$(az ad signed-in-user show --query id -o tsv)" --role "Key Vault Secrets Officer" --scope "$(az keyvault show --name ${vaultName} --resource-group ${installId}-rg --query id -o tsv)"
            `}</Code>
          </Card>

            {requiredSecrets?.map(renderSecretCard)}
              {overridableSecrets && overridableSecrets.length > 0 && (
                <Expand
                  id="overridable-secrets"
                  heading={
                    <Text variant="subtext">
                      Optional overrides ({overridableSecrets.length})
                    </Text>
                  }
                >
                  <div className="flex flex-col gap-4 p-2">
                    <Text variant="subtext">
                      These secrets have default values. Set them only if you need
                      to override the defaults.
                    </Text>
                    {overridableSecrets.map(renderSecretCard)}
                  </div>
                </Expand>
              )}
        </div>
      )}

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Deploy the install stack
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Preview changes (dry-run)</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={`az deployment group what-if --resource-group ${installId}-rg --template-uri ${stack?.versions?.at(0)?.template_url}`}
            />
          </span>
          <Code>{`
            az deployment group what-if --resource-group ${installId}-rg --template-uri ${stack?.versions?.at(0)?.template_url}
          `}</Code>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Deploy the stack to the resource group</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={`az stack group create --name ${installId}-stack --resource-group ${installId}-rg --template-uri ${stack?.versions?.at(0)?.template_url} --deny-settings-mode "denyDelete" --aou deleteAll`}
            />
          </span>
          <Code>{`
            az stack group create --name ${installId}-stack --resource-group ${installId}-rg --template-uri ${stack?.versions?.at(0)?.template_url} --deny-settings-mode "denyDelete" --aou deleteAll
          `}</Code>
        </Card>
      </div>

      <Divider dividerWord="or" />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Download the install stack template
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Install template link</Text>
            <ClickToCopyButton
              textToCopy={stack?.versions?.at(0)?.template_url}
            />
          </span>
          <Link
            href={stack?.versions?.at(0)?.template_url}
            target="_blank"
            rel="noopener noreferrer"
          >
            <Code>{stack?.versions?.at(0)?.template_url}</Code>
          </Link>
        </Card>
      </div>
    </>
  )
}

export const AwaitAzureDetailsSkeleton = () => {
  return (
    <>
      <Skeleton height="24px" width="175px" />

      <Card>
        <Skeleton height="17px" width="100px" />
        <Skeleton height="52px" width="100%" />
      </Card>

      <Card>
        <Skeleton height="17px" width="120px" />
        <Skeleton height="52px" width="100%" />
      </Card>

      <Divider dividerWord="or" />

      <Skeleton height="24px" width="325px" />

      <Card>
        <Skeleton height="17px" width="219px" />
        <Skeleton height="72px" width="100%" />
      </Card>
    </>
  )
}
