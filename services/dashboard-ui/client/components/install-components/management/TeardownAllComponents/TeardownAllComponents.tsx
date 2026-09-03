import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface ITeardownAllComponentsModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  isPending: boolean
  isKickedOff: boolean
  error?: { error?: string } | null
  onSubmit: () => void
}

export const TeardownAllComponentsModal = ({
  installName,
  isPending,
  isKickedOff,
  error,
  onSubmit,
  ...props
}: ITeardownAllComponentsModal) => {
  const [confirmName, setConfirmName] = useState('')
  const isConfirmValid = confirmName === installName

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
          <Icon variant="CloudArrowDownIcon" size="24" />
          Teardown all {installName} components?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting teardown
          </span>
        ) : (
          'Teardown all components'
        ),
        disabled: !isConfirmValid || isKickedOff || isPending,
        onClick: onSubmit,
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6 mb-6">
        {error?.error ? (
          <Banner theme="error">
            {error?.error || 'Unable to teardown components'}
          </Banner>
        ) : null}
        <Text variant="body" theme="neutral">
          This removes all running component deployments from {installName}.
        </Text>

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
            errorMessage={
              confirmName.length > 0 && !isConfirmValid
                ? "Install name doesn't match"
                : undefined
            }
          />
        </div>
      </div>
    </Modal>
  )
}
