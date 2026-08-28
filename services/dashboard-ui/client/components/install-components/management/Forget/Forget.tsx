import { type ReactNode, useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IForgetComponentModal extends IModal {
  componentName: string
  isLoading: boolean
  error: any
  onConfirm: () => void
  isTornDown?: boolean
  isInConfig?: boolean
  isConfigLoading?: boolean
  teardownAction?: ReactNode
}

const ChecklistItem = ({
  met,
  title,
  children,
}: {
  met: boolean
  title: string
  children?: ReactNode
}) => (
  <div className="flex items-start gap-3">
    {met ? (
      <Icon
        variant="CheckCircleIcon"
        size={18}
        weight="fill"
        className="text-green-600 dark:text-green-500 shrink-0 mt-0.5"
      />
    ) : (
      <span className="w-[18px] h-[18px] rounded-full border-2 border-cool-grey-300 dark:border-dark-grey-500 shrink-0 mt-0.5" />
    )}
    <div className="flex flex-col gap-1 min-w-0">
      <Text variant="body" weight={met ? 'strong' : 'stronger'}>
        {title}
      </Text>
      {!met && children ? children : null}
    </div>
  </div>
)

export const ForgetComponentModal = ({
  componentName,
  isLoading,
  error,
  onConfirm,
  isTornDown = true,
  isInConfig = false,
  isConfigLoading = false,
  teardownAction,
  ...props
}: IForgetComponentModal) => {
  const [confirmName, setConfirmName] = useState('')
  const isConfirmValid = confirmName === componentName

  const prerequisitesMet = isTornDown && !isInConfig && !isConfigLoading
  const canForget = prerequisitesMet && isConfirmValid && !isLoading

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
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
            removed. This removes {componentName} and cannot be undone.
          </Text>
        </Banner>

        <div className="flex flex-col gap-3">
          <Text variant="body" weight="strong">
            Before you can forget this component
          </Text>
          <div className="flex flex-col gap-4 border rounded-lg p-4">
            <ChecklistItem met={isTornDown} title="Component torn down">
              <Text variant="subtext" theme="neutral">
                Teardown the component so its infrastructure is removed from the
                cloud account.
              </Text>
              {teardownAction ? (
                <div className="mt-1">{teardownAction}</div>
              ) : null}
            </ChecklistItem>

            <ChecklistItem
              met={!isInConfig && !isConfigLoading}
              title="Removed from app config"
            >
              <Text variant="subtext" theme="neutral">
                Remove the component from your config, then run{' '}
                <Code variant="inline">nuon apps sync</Code> to update the app
                configuration.
              </Text>
            </ChecklistItem>
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <Text variant="body">
            To confirm, type{' '}
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
            disabled={!prerequisitesMet}
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
    </Modal>
  )
}
