import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { deleteImageRepository } from '@/lib'
import type { TImageRepository } from '@/types'

export const RemoveImageRepository = ({
  repository,
  ...props
}: { repository: TImageRepository } & IModal) => {
  const client = useQueryClient()
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const remove = useMutation({
    mutationFn: () =>
      deleteImageRepository({
        orgId: repository.org_id,
        repositoryId: repository.id,
      }),
    onSuccess: () => {
      client.removeQueries({
        queryKey: ['image-repositories', repository.org_id, repository.id],
      })
      client.invalidateQueries({
        queryKey: ['image-repositories', repository.org_id],
      })
      client.invalidateQueries({
        queryKey: ['cloud-connections', repository.org_id],
      })
      removeModal(props.modalId)
      navigate(`/${repository.org_id}/settings/image-repositories`)
      addToast(
        <Toast heading="Repository removed" theme="success">
          <Text>
            The repository assignment was removed. Cloud permissions are
            unchanged.
          </Text>
        </Toast>
      )
    },
  })
  return (
    <Modal
      {...props}
      heading="Remove image repository?"
      primaryActionTrigger={{
        variant: 'danger',
        children: remove.isPending
          ? 'Removing repository'
          : 'Remove repository',
        disabled: remove.isPending,
        onClick: () => remove.mutate(),
      }}
    >
      <div className="flex flex-col gap-4">
        <Text>
          Removing <Text family="mono">{repository.image_url_prefix}</Text>{' '}
          deletes its saved pull identity selection. It does not delete images,
          the identity, or cloud permissions.
        </Text>
        <Banner theme="warn">
          New config saves and app syncs for this repository will fail until it
          is registered again. Existing configurations keep their pinned
          identity. To change identities, add the repository again with the
          replacement and re-sync the app.
        </Banner>
        <FormErrorBanner
          error={remove.error}
          fallback="Repository removal failed"
        />
      </div>
    </Modal>
  )
}
