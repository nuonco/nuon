'use client'

import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'

interface IDeprovisionStack {}

export const DeprovisionStackModal = ({ ...props }: IDeprovisionStack & IModal) => {
  const { removeModal } = useSurfaces()
  const { install } = useInstall()

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="warn"
        >
          <Icon variant="StackMinus" size="24" />
          Deprovision stack for {install.name}?
        </Text>
      }
      primaryActionTrigger={{
        children: 'Got it',
        onClick: () => removeModal(props.modalId),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <Banner theme="warn">
          <Text variant="body">
            <strong>Manual Action Required:</strong> Once you have deprovisioned the install from the UI, please go to the cloud platform console and destroy this stack for your install.
          </Text>
        </Banner>

        <div className="flex flex-col gap-3">
          <Text variant="body" weight="strong">
            How to deprovision the CloudFormation stack:
          </Text>
          <ul className="flex flex-col gap-2 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
            <li>Navigate to the AWS CloudFormation console in your account</li>
            <li>Find the stack associated with this install: <span className="font-mono bg-cool-grey-100 dark:bg-cool-grey-800 px-1 py-0.5 rounded">{install.name}</span></li>
            <li>Select the stack and click &quot;Delete&quot;</li>
            <li>Confirm the deletion to remove all associated resources</li>
          </ul>
        </div>

        <Banner theme="info">
          <Text variant="body">
            <strong>Note:</strong> This action must be performed manually in the AWS console. The UI cannot automatically delete CloudFormation stacks in your account.
          </Text>
        </Banner>
      </div>
    </Modal>
  )
}

export const DeprovisionStackButton = ({
  ...props
}: IDeprovisionStack & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <DeprovisionStackModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Deprovision stack
      <Icon variant="StackMinus" />
    </Button>
  )
}