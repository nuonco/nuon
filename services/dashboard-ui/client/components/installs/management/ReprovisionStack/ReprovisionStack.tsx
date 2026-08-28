import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IReprovisionStackModal extends Omit<IModal, 'onSubmit'> {
  installId: string
  installName: string
  isPending: boolean
  error: any
  onSubmit: (params: { selectedRole: string; skipComponents: boolean }) => void
  onClose: () => void
}

export const ReprovisionStackModal = ({
  installId,
  installName,
  isPending,
  error,
  onSubmit,
  onClose,
  ...props
}: IReprovisionStackModal) => {
  const [selectedRole, setSelectedRole] = useState<string>('')
  // Defaults to skipping components: a stack reprovision recreates the runner, not
  // what is running on the sandbox.
  const [skipComponents, setSkipComponents] = useState(true)

  return (
    <Modal
      heading="Reprovision stack?"
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Reprovisioning
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="StackPlusIcon" />
            Reprovision stack
          </span>
        ),
        disabled: isPending,
        onClick: () => onSubmit({ selectedRole, skipComponents }),
        variant: 'primary' as const,
      }}
      onClose={onClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to reprovision stack.'}
          </Banner>
        ) : null}

        <Text variant="body" className="leading-relaxed">
          Reprovisioning will recreate the stack and runner for {installName}.
        </Text>

        <Banner theme="warn">
          <Text variant="body">
            <strong>Warning:</strong> Actions and deployments won't be available
            while the runner is recreated.
          </Text>
        </Banner>

        <RoleSelector
          installId={installId}
          operationType="reprovision"
          value={selectedRole}
          onChange={setSelectedRole}
          name="role"
        />

        <div className="flex items-start">
          <CheckboxInput
            checked={skipComponents}
            onChange={(e) => setSkipComponents(e.target.checked)}
            className="mt-1.5"
            labelProps={{
              className:
                'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !p-2 gap-4 max-w-none !items-start',
              labelText: (
                <div className="flex flex-col gap-1">
                  <Text variant="base" weight="stronger">
                    Skip component deployments
                  </Text>
                  <Text variant="subtext" theme="neutral">
                    Only reprovision the stack without redeploying components on
                    top.
                  </Text>
                </div>
              ),
            }}
          />
        </div>
      </div>
    </Modal>
  )
}
