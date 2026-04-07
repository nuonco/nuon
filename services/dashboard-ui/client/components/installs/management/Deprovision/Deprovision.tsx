import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IDeprovisionModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  isPending: boolean
  error: any
  onSubmit: () => void
}

export const DeprovisionModal = ({
  installName,
  isPending,
  error,
  onSubmit,
  ...props
}: IDeprovisionModal) => {
  const [confirmName, setConfirmName] = useState('')

  const isConfirmValid = confirmName === installName
  const canDeprovision = isConfirmValid && !isPending

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="ArrowDown" size="24" />
          Deprovision entire install
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Deprovisioning...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="ArrowDown" />
            Deprovision install
          </span>
        ),
        onClick: onSubmit,
        disabled: !canDeprovision,
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to kick off install deprovision'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Are you sure you want to deprovision {installName}?
            </Text>
            <Text variant="body" theme="neutral">
              Deprovisioning an install will remove it from the cloud account.
            </Text>
          </div>

          <div className="flex flex-col gap-3">
            <Text variant="body">
              This will create a workflow that attempts to:
            </Text>
            <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
              <li>Teardown each install component according to the dependency order.</li>
              <li>Teardown the install sandbox</li>
            </ul>
          </div>

          <div className="flex flex-col gap-2">
            <Text variant="body">
              To verify, type{' '}
              <span className="font-mono font-medium text-red-800 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-1 py-0.5 rounded">
                {installName}
              </span>{' '}
              below.
            </Text>
            <Input
              id="confirm-install-name"
              placeholder="install name"
              type="text"
              value={confirmName}
              onChange={(e) => setConfirmName(e.target.value)}
              error={confirmName.length > 0 && !isConfirmValid}
              errorMessage={confirmName.length > 0 && !isConfirmValid ? "Install name doesn't match" : undefined}
            />
          </div>

          <Banner theme="warn">
            <Text variant="body">
              <strong>Important:</strong> After this workflow completes, please manually teardown the CloudFormation stack in the AWS console.
            </Text>
          </Banner>
        </div>
      </div>
    </Modal>
  )
}
