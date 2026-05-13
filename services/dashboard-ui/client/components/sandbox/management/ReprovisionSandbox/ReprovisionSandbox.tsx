import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IReprovisionSandboxModal extends Omit<IModal, 'onSubmit'> {
  installId: string
  isPending: boolean
  error: any
  onSubmit: (params: { selectedRole: string; skipComponents: boolean }) => void
  onClose: () => void
}

export const ReprovisionSandboxModal = ({
  installId,
  isPending,
  error,
  onSubmit,
  onClose,
  ...props
}: IReprovisionSandboxModal) => {
  const [selectedRole, setSelectedRole] = useState<string>('')
  const [skipComponents, setSkipComponents] = useState(false)

  return (
    <Modal
      heading="Reprovision sandbox?"
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Reprovisioning
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="BoxArrowUpIcon" />
            Reprovision sandbox
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
            {(error as any)?.error || 'Unable to reprovision sandbox.'}
          </Banner>
        ) : null}

        <Text variant="body" className="leading-relaxed">
          Are you sure you want to reprovision this sandbox?
        </Text>

        <RoleSelector
          installId={installId}
          operationType="reprovision"
          principalType="sandbox"
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
                    Only reprovision the sandbox infrastructure without redeploying components on top.
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
