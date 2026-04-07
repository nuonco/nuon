import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IDeprovisionStackModal extends IModal {
  installName: string
  onDismiss: () => void
}

export const DeprovisionStackModal = ({
  installName,
  onDismiss,
  ...props
}: IDeprovisionStackModal) => {
  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="warn"
        >
          <Icon variant="StackMinus" size="24" />
          Deprovision stack for {installName}?
        </Text>
      }
      primaryActionTrigger={{
        children: 'Got it',
        onClick: onDismiss,
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
            <li>Find the stack associated with this install: <span className="font-mono bg-cool-grey-100 dark:bg-cool-grey-800 px-1 py-0.5 rounded">{installName}</span></li>
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
