import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TComponent } from '@/types'

interface IToggleComponentModal extends Omit<IModal, 'onSubmit'> {
  component: TComponent
  enabling: boolean
  isPending: boolean
  error?: { error?: string } | null
  onSubmit: (params: { role: string }) => void
  onClose: () => void
  roleSelector: (props: {
    value: string
    onChange: (value: string) => void
  }) => ReactNode
}

export const ToggleComponentModal = ({
  component,
  enabling,
  isPending,
  error,
  onSubmit,
  onClose,
  roleSelector,
  ...props
}: IToggleComponentModal) => {
  const action = enabling ? 'Enable' : 'Disable'
  const actionLower = enabling ? 'enable' : 'disable'
  const activeAction = enabling ? 'Enabling' : 'Disabling'
  const iconVariant = enabling ? 'ToggleRightIcon' : 'ToggleLeftIcon'

  return (
    <Modal
      heading={`${action} ${component.name}?`}
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            {activeAction}
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant={iconVariant} />
            {action} component
          </span>
        ),
        disabled: isPending,
        onClick: () => onSubmit({ role: '' }),
        variant: enabling ? ('primary' as const) : ('danger' as const),
      }}
      onClose={onClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || `Unable to ${actionLower} component`}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          {enabling ? (
            <Text variant="body" theme="neutral">
              Enabling {component.name} will deploy this component to the
              install. A plan and apply workflow will be created.
            </Text>
          ) : (
            <Text variant="body" theme="neutral">
              Disabling {component.name} will tear down all infrastructure for
              this component. A teardown workflow will be created.
            </Text>
          )}

          {roleSelector({
            value: '',
            onChange: () => {},
          })}
        </div>
      </div>
    </Modal>
  )
}
