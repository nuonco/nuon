import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { deleteCloudConnection } from '@/lib'
import type { TCloudConnection } from '@/types'
import { DeleteConnection } from './DeleteConnection'

export const DeleteConnectionContainer = ({
  connection,
  ...props
}: { connection: TCloudConnection } & IModal) => {
  const client = useQueryClient()
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const mutation = useMutation({
    mutationFn: () =>
      deleteCloudConnection({
        orgId: connection.org_id,
        connectionId: connection.id,
      }),
    onSuccess: () => {
      navigate(`/${connection.org_id}/settings/cloud-connections`)
      client.removeQueries({
        queryKey: ['cloud-connections', connection.org_id, connection.id],
      })
      client.invalidateQueries({
        queryKey: ['cloud-connections', connection.org_id],
      })
      removeModal(props.modalId)
      addToast(
        <Toast heading="Connection deleted" theme="success">
          <Text>{connection.name} was deleted.</Text>
        </Toast>
      )
    },
  })
  return (
    <DeleteConnection
      {...props}
      name={connection.name}
      error={mutation.error}
      isPending={mutation.isPending}
      onDelete={() => mutation.mutate()}
    />
  )
}
