import { useState } from 'react'
import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TComponent } from '@/types'

interface ITeardownComponentModal extends Omit<IModal, 'onSubmit'> {
  component: TComponent
  isPending: boolean
  error?: { error?: string } | null
  onSubmit: (params: { role: string }) => void
  onClose: () => void
  roleSelector: (props: {
    value: string
    onChange: (value: string) => void
  }) => ReactNode
}

export const TeardownComponentModal = ({
  component,
  isPending,
  error,
  onSubmit,
  onClose,
  roleSelector,
  ...props
}: ITeardownComponentModal) => {
  const [confirmName, setConfirmName] = useState('')
  const [selectedRole, setSelectedRole] = useState<string>('')

  const isConfirmValid = confirmName === component.name
  const canTeardown = isConfirmValid && !isPending

  const handleClose = () => {
    setConfirmName('')
    onClose()
  }

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            flex
            className="gap-4"
            variant="h3"
            weight="strong"
            theme="error"
          >
            <Icon variant="Warning" size="24" />
            Teardown {component.name} component
          </Text>
          <Text
            variant="body"
            className="text-cool-grey-600 dark:text-cool-grey-400"
          >
            This will remove all running deployments for this component
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Tearing down
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="CloudArrowDownIcon" />
            Teardown component
          </span>
        ),
        disabled: !canTeardown,
        onClick: () => onSubmit({ role: selectedRole }),
        variant: 'danger' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to teardown component'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Are you sure you want to teardown {component.name}?
            </Text>
            <Text variant="body" theme="neutral">
              Tearing down a component will remove all of its running deployments from the cloud account.
            </Text>
          </div>

          <div className="flex flex-col gap-3">
            <Text variant="body">
              This will create a workflow that attempts to:
            </Text>
            <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
              <li>Remove all infrastructure provisioned by this component</li>
              <li>Clean up all associated resources and dependencies</li>
            </ul>
          </div>

          <div className="flex flex-col gap-2">
            <Text variant="body">
              To verify, type{' '}
              <span className="font-mono font-medium text-red-800 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-1 py-0.5 rounded">
                {component.name}
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
              errorMessage={confirmName.length > 0 && !isConfirmValid ? "Component name doesn't match" : undefined}
            />
          </div>

          <Banner theme="warn">
            <Text variant="body">
              <strong>Important:</strong> This action cannot be undone. All infrastructure provisioned by this component will be permanently destroyed.
            </Text>
          </Banner>

          {roleSelector({
            value: selectedRole,
            onChange: setSelectedRole,
          })}
        </div>
      </div>
    </Modal>
  )
}
