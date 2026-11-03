import { Banner } from '@/components/common/Banner'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const DeleteConnection = ({
  name,
  orgId,
  imageRepositories = 0,
  error,
  isPending,
  onDelete,
  ...props
}: {
  name: string
  orgId?: string
  imageRepositories?: number
  error?: TAPIError | null
  isPending: boolean
  onDelete: () => void
} & IModal) => (
  <Modal
    {...props}
    heading="Delete cloud connection?"
    primaryActionTrigger={{
      variant: 'danger',
      children: isPending ? 'Deleting connection' : 'Delete connection',
      disabled: isPending || imageRepositories > 0,
      tooltipProps:
        imageRepositories > 0
          ? {
              tipContent:
                'Cannot delete — remove image repository assignments first',
            }
          : undefined,
      onClick: onDelete,
    }}
  >
    <div className="flex flex-col gap-4">
      <Text>
        Deleting {name} removes this connection from Nuon. It does not delete
        the principal or any cloud resources.
      </Text>
      {imageRepositories > 0 && (
        <Banner theme="warn">
          This connection is assigned to {imageRepositories} image{' '}
          {imageRepositories === 1 ? 'repository' : 'repositories'}. Remove{' '}
          {imageRepositories === 1 ? 'that assignment' : 'those assignments'} in{' '}
          {orgId ? (
            <Link href={`/${orgId}/settings/image-repositories`}>
              Image repositories
            </Link>
          ) : (
            'Image repositories'
          )}{' '}
          before deleting it.
        </Banner>
      )}
      <FormErrorBanner
        error={
          error
            ? {
                ...error,
                error: 'Cannot delete connection',
                description: error.description || error.error,
              }
            : error
        }
        fallback="Cannot delete connection"
      />
    </div>
  </Modal>
)
