'use client'

import { Button, type IButtonAsButton } from '@/components/common/Button'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useQuery } from '@/hooks/use-query'
import { useWorkflowActions } from '@/hooks/use-workflow-actions'
import type { TWorkflow, TInstallStack } from '@/types'

interface ICustomerShare {
  workflow: TWorkflow
}

export const CustomerShareModal = ({
  workflow,
  ...props
}: ICustomerShare & IModal) => {
  const { install } = useInstall()

  const { data: stack, isLoading } = useQuery<TInstallStack>({
    path: `/api/orgs/${install?.org_id}/installs/${install?.id}/stack`,
  })

  console.log('stack', stack)

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="Export" size="24" />
          {`${install?.app?.name} install link`}
        </Text>
      }
      size="half"
      {...props}
    >
      <div className="flex flex-col gap-1">
        <Text variant="base" weight="stronger">
          {`Share the following message with your customer to install ${install?.app?.name} app in their cloud account.`}
        </Text>

        {isLoading ? (
          <div className="flex flex-col gap-2 mt-4">
            <Skeleton height="17px" width="100px" />
            <Skeleton height="132px" width="100%" />
          </div>
        ) : stack?.versions?.at(0)?.quick_link_url ? (
          <>
            <span className="flex justify-between items-center mt-4">
              <Text weight="strong">Install quick link</Text>
              <ClickToCopyButton
                textToCopy={stack?.versions?.at(0)?.quick_link_url}
              />
            </span>
            <Link
              href={stack?.versions?.at(0)?.quick_link_url}
              target="_blank"
              rel="noopener noreferrer"
            >
              <Code className="!my-0">
                {stack?.versions?.at(0)?.quick_link_url}
              </Code>
            </Link>
          </>
        ) : null}

        <Text>
          {`This is a Cloudformation stack that will install ${install?.app?.name}.`}
        </Text>

        <Text className="mt-4">To use it, follow these steps:</Text>
        <ol className="mb-4">
          <li>
            <Text>1. Log into the account you want to install the app in.</Text>
          </li>
          <li>
            <Text>
              2. Click on the link. It will take you to the AWS Cloudformation
              console.
            </Text>
          </li>
          <li>
            <Text>
              3. Fill out the inputs and secrets, then follow the stack install
              directions.
            </Text>
          </li>
          <li>
            <Text>
              {`4. When the stack has provisioned, it will automatically install ${install?.app?.name}.`}
            </Text>
          </li>
        </ol>
        <Text>
          {`The provider will be automatically notified when the stack has provisioned, and will monitor the installation of ${install?.app?.name}.`}
        </Text>
      </div>
    </Modal>
  )
}

export const CustomerShareButton = ({
  workflow,
  ...props
}: ICustomerShare & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <CustomerShareModal workflow={workflow} />
  const { canShowCustomerShare } = useWorkflowActions(workflow, false)

  return canShowCustomerShare ? (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Share with customer
      {props?.isMenuButton ? <Icon variant="Export" /> : null}
    </Button>
  ) : null
}
