import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IDeprovisionSandboxModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  installId: string
  isPending: boolean
  error: any
  onSubmit: (params: { selectedRole: string }) => void
  onClose: () => void
}

export const DeprovisionSandboxModal = ({
  installName,
  installId,
  isPending,
  error,
  onSubmit,
  onClose,
  ...props
}: IDeprovisionSandboxModal) => {
  const [confirm, setConfirm] = useState<string>('')
  const [selectedRole, setSelectedRole] = useState<string>('')

  return (
    <Modal
      className="!max-w-xl"
      heading="Deprovision install sandbox"
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Deprovisioning
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="BoxArrowDownIcon" />
            Deprovision sandbox
          </span>
        ),
        disabled: confirm !== 'deprovision' || isPending,
        onClick: () => onSubmit({ selectedRole }),
        variant: 'danger' as const,
      }}
      onClose={onClose}
      {...props}
    >
      <div className="flex flex-col gap-8 mb-12">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to kickoff sandbox deprovision'}
          </Banner>
        ) : null}

        <span className="flex flex-col gap-1">
          <Text variant="h3" weight="strong">
            Are you sure you want to deprovision {installName} sandbox?
          </Text>
          <Text
            className="text-cool-grey-600 dark:text-white/70"
            variant="subtext"
          >
            Deprovisioning a sandbox will remove it from the cloud account.
          </Text>
        </span>

        <div className="flex flex-col gap-2">
          <Text variant="body">
            This will create a workflow that attempts to:
          </Text>

          <ul className="flex flex-col gap-1 list-disc pl-4">
            <li className="text-sm">Teardown the install sandbox</li>
          </ul>
        </div>

        <div className="w-full">
          <label className="flex flex-col gap-1 w-full">
            <Text variant="base" weight="strong">
              To verify, type{' '}
              <span className="text-red-800 dark:text-red-500">
                deprovision
              </span>{' '}
              below.
            </Text>
            <Input
              placeholder="deprovision"
              className="w-full"
              type="text"
              value={confirm}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                setConfirm(e?.currentTarget?.value)
              }}
            />
          </label>
        </div>

        <RoleSelector
          installId={installId}
          operationType="deprovision"
          principalType='sandbox'
          value={selectedRole}
          onChange={setSelectedRole}
          name="role"
        />
      </div>
    </Modal>
  )
}
