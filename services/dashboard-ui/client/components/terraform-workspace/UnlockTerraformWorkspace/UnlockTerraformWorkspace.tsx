import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

interface IUnlockTerraformWorkspaceModal extends Omit<IModal, 'onSubmit'> {
  description?: string
  isPending: boolean
  error?: TAPIError | null
  onSubmit: () => void
  onClose: () => void
}

export const UnlockTerraformWorkspaceModal = ({
  description = 'the workspace',
  isPending,
  error,
  onSubmit,
  onClose,
  ...props
}: IUnlockTerraformWorkspaceModal) => {
  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text flex className="gap-4" variant="h3" weight="strong">
            <Icon variant="LockOpen" size="24" />
            Unlock Terraform workspace
          </Text>
          <Text variant="body" className="text-cool-grey-600 dark:text-cool-grey-400">
            Force unlock the Terraform state for {description}
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Unlocking...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="LockOpen" />
            Force unlock
          </span>
        ),
        disabled: isPending,
        onClick: onSubmit,
        variant: 'danger' as const,
      }}
      onClose={onClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to unlock Terraform workspace'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <Text variant="body" theme="neutral">
            Are you sure you want to force unlock this Terraform workspace? This should only be
            done if a previous operation failed to release the lock.
          </Text>

          <Banner theme="warn">
            <Text variant="body">
              <strong>Warning:</strong> Force unlocking a workspace that is actively in use by a
              running job may cause state corruption.
            </Text>
          </Banner>
        </div>
      </div>
    </Modal>
  )
}
