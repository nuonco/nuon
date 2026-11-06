import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { deleteImageRepository } from '@/lib'
import type { TImageRepository } from '@/types'
import { RemoveImageRepository } from './RemoveImageRepository'

export const RemoveImageRepositoryContainer = ({
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
            Removed <Text family="mono">{repository.image_url_prefix}</Text>.
            Cloud permissions are unchanged.
          </Text>
        </Toast>
      )
    },
  })
  return (
    <RemoveImageRepository
      {...props}
      name={repository.image_url_prefix}
      error={remove.error}
      isPending={remove.isPending}
      onDelete={() => remove.mutate()}
    />
  )
}
