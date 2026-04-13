import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { IStackDetails } from '../types'

interface IAwaitAzureDetails extends IStackDetails {
  installId: string
  azureLocation?: string
}

export const AwaitAzureDetails = ({ stack, installId, azureLocation }: IAwaitAzureDetails) => {
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
