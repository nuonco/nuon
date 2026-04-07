import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IForgetModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  isPending: boolean
  error: any
  onSubmit: () => void
}

export const ForgetModal = ({
  installName,
  isPending,
  error,
  onSubmit,
  ...props
}: IForgetModal) => {
  const [confirmName, setConfirmName] = useState('')

  const isConfirmValid = confirmName === installName
  const canForget = isConfirmValid && !isPending

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
          <Icon variant="Trash" size="24" />
          Forget {installName}
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Forgetting...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="Trash" />
            Forget install
          </span>
        ),
        onClick: onSubmit,
        disabled: !canForget,
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to forget install.'}
          </Banner>
        ) : null}

        <Banner theme="warn">
          <Text variant="body">
            <strong>Warning:</strong> This should only be used in cases where an install was broken in an unordinary way and needs to be manually removed.
          </Text>
        </Banner>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Are you sure you want to forget {installName}?
            </Text>
            <Text variant="body" theme="neutral">
              This action will remove the install and can not be undone.
            </Text>
          </div>

          <div className="flex flex-col gap-3">
            <Text variant="body">
              You should only do this after you have:
            </Text>
            <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
              <li>Successfully deprovisioned the install</li>
              <li>Deprovisioned the CloudFormation stack for this install</li>
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
        </div>
      </div>
    </Modal>
  )
}
