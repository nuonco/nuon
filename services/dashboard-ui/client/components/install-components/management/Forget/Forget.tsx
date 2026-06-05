import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IForgetComponentModal extends IModal {
  componentName: string
  isLoading: boolean
  error: any
  onConfirm: () => void
}

export const ForgetComponentModal = ({
  componentName,
  isLoading,
  error,
  onConfirm,
  ...props
}: IForgetComponentModal) => {
  const [confirmName, setConfirmName] = useState('')
  const isConfirmValid = confirmName === componentName
  const canForget = isConfirmValid && !isLoading

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
          <Icon variant="TrashIcon" size="24" />
          Forget {componentName}?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Forgetting...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="TrashIcon" />
            Forget component
          </span>
        ),
        onClick: onConfirm,
        disabled: !canForget,
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error?.error ? (
          <Banner theme="error">
            {error?.error || 'Unable to forget component.'}
          </Banner>
        ) : null}

        <Banner theme="warn">
          <Text variant="body">
            <strong>Warning:</strong> This should only be used in cases where a
            component was broken in an unordinary way and needs to be manually
            removed.
          </Text>
        </Banner>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Text variant="body" theme="neutral">
              This will remove {componentName} and cannot be undone.
            </Text>
          </div>

          <div className="flex flex-col gap-3">
            <Text variant="body">You should only do this after you have:</Text>
            <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
              <li>Successfully tore down the component</li>
              <li>Verified no infrastructure remains in the cloud account</li>
              <li>Confirmed all dependencies are handled</li>
            </ul>
          </div>

          <div className="flex flex-col gap-2">
            <Text variant="body">
              To verify, type{' '}
              <span className="font-mono font-medium text-red-800 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-1 py-0.5 rounded">
                {componentName}
              </span>{' '}
              below.
            </Text>
            <Input
              id="confirm-component-name"
              placeholder="component name"
              type="text"
              value={confirmName}
              onChange={(e) => setConfirmName(e.target.value)}
              error={confirmName.length > 0 && !isConfirmValid}
              errorMessage={
                confirmName.length > 0 && !isConfirmValid
                  ? "Component name doesn't match"
                  : undefined
              }
            />
          </div>
        </div>
      </div>
    </Modal>
  )
}
