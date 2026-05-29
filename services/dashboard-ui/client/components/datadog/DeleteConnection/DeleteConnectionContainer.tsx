import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { deleteDatadogConnection } from '@/lib'
import type { TAPIError, TDatadogConnection } from '@/types'
import { DeleteConnectionModal } from './DeleteConnection'

const DeleteConnectionModalContainer = (
  props: { connection: TDatadogConnection } & Record<string, any>
) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending } = useMutation({
    mutationFn: () =>
      deleteDatadogConnection({
        orgId: org.id,
        connectionId: props.connection.id!,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['datadog-connections', org.id],
      })
      queryClient.invalidateQueries({
        queryKey: ['datadog-event-subscriptions', org.id],
      })
      queryClient.invalidateQueries({
        queryKey: ['datadog-managed-monitors', org.id],
      })
      addToast(<Toast heading="Connection deleted" theme="success" />)
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to delete connection" theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <DeleteConnectionModal
      connection={props.connection}
      isPending={isPending}
      onConfirm={() => mutate()}
      {...props}
    />
  )
}

export const DeleteConnectionButton = ({
  connection,
  ...props
}: { connection: TDatadogConnection } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <DeleteConnectionModalContainer connection={connection} />

  return (
    <Button variant="ghost" onClick={() => addModal(modal)} {...props}>
      <Icon variant="TrashIcon" size={14} />
    </Button>
  )
}
