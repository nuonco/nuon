'use client'

import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'

import { useWorkflowActions } from '@/hooks/use-workflow-actions'
import type { TWorkflow } from '@/types'

interface ICustomerShare {
  workflow: TWorkflow
}

export const CustomerShareModal = ({
  workflow,
  ...props
}: ICustomerShare & IModal) => {
  const { install } = useInstall()

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
      {...props}
    >
      <div className="flex flex-col gap-1">
        <Text variant="base" weight="stronger">
          {`Share the following message with your customer to install ${install?.app?.name} app in their cloud account.`}
        </Text>
        <Text variant="base">(install link goes here)</Text>
        <Text variant="base">
          {`This is a Cloudformation stack that will install ${install?.app?.name}. To use it, follow these steps:`}
        </Text>
        <Text variant="base">
          1. Log into the account you want to install the app in.
        </Text>
        <Text variant="base">
          2. Click on the link. It will take you to the AWS Cloudformation
          console.
        </Text>
        <Text variant="base">
          3. Fill out the inputs and secrets, then follow the stack install
          directions.
        </Text>
        <Text variant="base">
          {`4. When the stack has provisioned, it will automatically install ${install?.app?.name}.`}
        </Text>
        <Text variant="base">
          {`(name_of_vendor) will be automatically notified when the stack has provisioned, and will monitor the installation of ${install?.app?.name}.`}
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
