import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TStackType } from '@/types'

interface IStackCopy {
  console: string
  stackNoun: string
  steps: (installName: string) => string[]
  note: string
}

const STACK_COPY: Record<TStackType, IStackCopy> = {
  'aws-cloudformation': {
    console: 'AWS CloudFormation console',
    stackNoun: 'CloudFormation stack',
    steps: (installName) => [
      'Navigate to the AWS CloudFormation console in your account',
      `Find the stack associated with this install: ${installName}`,
      'Select the stack and click "Delete"',
      'Confirm the deletion to remove all associated resources',
    ],
    note: 'This action must be performed manually in the AWS console. The UI cannot automatically delete CloudFormation stacks in your account.',
  },
  'azure-bicep': {
    console: 'Azure portal',
    stackNoun: 'Azure deployment stack',
    steps: (installName) => [
      'Navigate to the deployment stacks section of the Azure portal',
      `Find the deployment stack associated with this install: ${installName}`,
      'Delete the deployment stack, choosing to delete all managed resources',
      'Confirm the deletion in the Azure portal',
    ],
    note: 'This action must be performed manually in the Azure portal. The UI cannot automatically delete deployment stacks in your subscription.',
  },
  'gcp-terraform': {
    console: 'Google Cloud console',
    stackNoun: 'Infrastructure Manager deployment',
    steps: (installName) => [
      'Open Infrastructure Manager in the Google Cloud console for your project',
      `Find the deployment associated with this install: ${installName}`,
      'Delete the deployment to destroy all associated resources',
      'If you deployed the stack with Terraform directly, run terraform destroy instead',
    ],
    note: 'This action must be performed manually in Google Cloud. The UI cannot automatically destroy Infrastructure Manager deployments in your project.',
  },
}

const GENERIC_COPY: IStackCopy = {
  console: "your cloud provider's console",
  stackNoun: 'stack',
  steps: (installName) => [
    "Open your cloud provider's console for the account this install runs in",
    `Find the stack associated with this install: ${installName}`,
    'Delete the stack to remove all associated resources',
    'Confirm the deletion in your cloud provider',
  ],
  note: "This action must be performed manually in your cloud provider's console. The UI cannot automatically delete stacks in your account.",
}

interface IDeprovisionStackModal extends IModal {
  installName: string
  stackType?: TStackType
  onDismiss: () => void
}

export const DeprovisionStackModal = ({
  installName,
  stackType,
  onDismiss,
  ...props
}: IDeprovisionStackModal) => {
  const copy = (stackType && STACK_COPY[stackType]) || GENERIC_COPY
  const steps = copy.steps(installName)

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
          <Icon variant="StackMinusIcon" size="24" />
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
            <strong>Manual action required:</strong> Once you have deprovisioned the install from the UI, go to {copy.console} and destroy this stack for your install.
          </Text>
        </Banner>

        <div className="flex flex-col gap-3">
          <Text variant="body" weight="strong">
            How to deprovision the {copy.stackNoun}:
          </Text>
          <ul className="flex flex-col gap-2 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
            {steps.map((step, i) => {
              const [before, after] = step.split(installName)
              return (
                <li key={i}>
                  {after === undefined ? (
                    step
                  ) : (
                    <>
                      {before}
                      <span className="font-mono bg-cool-grey-100 dark:bg-cool-grey-800 px-1 py-0.5 rounded">{installName}</span>
                      {after}
                    </>
                  )}
                </li>
              )
            })}
          </ul>
        </div>

        <Banner theme="info">
          <Text variant="body">
            <strong>Note:</strong> {copy.note}
          </Text>
        </Banner>
      </div>
    </Modal>
  )
}
