import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IDisableConfigSyncModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  isPending: boolean
  error: any
  onSubmit: () => void
}

export const DisableConfigSyncModal = ({
  installName,
  isPending,
  error,
  onSubmit,
  ...props
}: IDisableConfigSyncModal) => {
  const [confirmName, setConfirmName] = useState('')

  const isConfirmValid = confirmName === installName
  const canSubmit = isConfirmValid && !isPending

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong" theme="warn">
          <Icon variant="FileCloudIcon" size="24" />
          Disable config sync?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Disabling...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="ToggleRightIcon" /> Disable config sync
          </span>
        ),
        onClick: onSubmit,
        disabled: !canSubmit,
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        {error ? (
          <Banner theme="error">
            {error?.error || 'An error happened, please refresh the page and try again.'}
          </Banner>
        ) : null}

        <Banner theme="warn">
          <Text variant="body">
            <strong>Warning:</strong> Disabling config sync will stop future syncs from the install config file. Settings will need to be managed manually from the dashboard.
          </Text>
        </Banner>

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
      </div>
    </Modal>
  )
}

interface IEnableConfigSyncModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  error: any
  onSubmit: () => void
}

export const EnableConfigSyncModal = ({
  isPending,
  error,
  onSubmit,
  ...props
}: IEnableConfigSyncModal) => {
  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="FileCloudIcon" size="24" />
          Enable config sync?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Enabling...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="ToggleLeftIcon" /> Enable config sync
          </span>
        ),
        onClick: onSubmit,
        disabled: isPending,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        {error ? (
          <Banner theme="error">
            {error?.error || 'An error happened, please refresh the page and try again.'}
          </Banner>
        ) : null}

        <Text variant="base">
          Enable this to allow syncing settings from an install config file.
        </Text>
      </div>
    </Modal>
  )
}
