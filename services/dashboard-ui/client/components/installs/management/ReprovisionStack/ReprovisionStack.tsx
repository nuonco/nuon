import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IReprovisionStackModal extends Omit<IModal, 'onSubmit'> {
  installId: string
  installName: string
  isPending: boolean
  error: any
  onSubmit: (params: { selectedRole: string }) => void
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
        onClick: () => onSubmit({ selectedRole }),
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
          Reprovisioning will recreate the stack and runner for {installName}. Components will not be redeployed.
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
      </div>
    </Modal>
  )
}
