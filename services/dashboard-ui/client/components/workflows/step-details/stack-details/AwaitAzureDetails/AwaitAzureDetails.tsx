import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
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
  // Azure stacks never create vault secrets — customers pre-create all of them, even auto-generated/defaulted ones
  const allSecrets = [...(secrets ?? [])].sort(
    (a, b) => Number(!!b.required) - Number(!!a.required)
  )

  const secretValue = (secret: TAppSecretConfig) => {
    if (secret.auto_generate) {
      return secret.format === 'base64'
        ? '$(openssl rand -base64 48)'
        : '$(openssl rand -hex 32)'
    }
    return secret.default || '<your-secret-value>'
  }

  const secretHint = (secret: TAppSecretConfig) => {
    if (secret.auto_generate) {
      return 'The command generates a random value for this secret.'
    }
    if (secret.default) {
      return 'The command pre-fills the app default. Replace the value to override it.'
    }
    if (!secret.required) {
      return 'Optional. Create it only if your app needs a value.'
    }
    return null
  }

  const renderSecretCard = (secret: TAppSecretConfig) => {
    const kvName = (secret.name ?? '').replaceAll('_', '-')
    const cmd = `az keyvault secret set --vault-name ${vaultName} --name ${kvName} --value "${secretValue(secret)}"`
    const hint = secretHint(secret)
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
        {hint && <Text variant="subtext">{hint}</Text>}
        {secret.format === 'base64' && !secret.auto_generate && (
          <Text variant="subtext">The value must be base64-encoded.</Text>
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

      {allSecrets.length > 0 && (
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Create secrets in the Key Vault
          </Text>
          <Text variant="subtext">
            Before deploying the stack, create every secret below in the Key
            Vault. The stack does not create secrets — a skipped secret is never
            synced, even when it has a default. The secret names must match
            exactly.
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

          {allSecrets.map(renderSecretCard)}
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
