import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { updateDatadogConnection } from '@/lib'
import type {
  TAPIError,
  TDatadogConnection,
  TUpdateDatadogConnectionBody,
} from '@/types'
import { EditConnectionModal } from './EditConnection'

const EditConnectionModalContainer = (
  props: { connection: TDatadogConnection } & Record<string, any>
) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: (body: TUpdateDatadogConnectionBody) =>
      updateDatadogConnection({
        orgId: org.id,
        connectionId: props.connection.id!,
        body,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['datadog-connections', org.id],
      })
      queryClient.invalidateQueries({
        queryKey: ['datadog-connection', org.id, props.connection.id],
      })
      addToast(
        <Toast heading="Connection updated" theme="success">
          <Text>Future events will use the updated settings.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to update connection" theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <EditConnectionModal
      connection={props.connection}
      isPending={isPending}
      error={error}
      onSubmit={mutate}
      {...props}
    />
  )
}

export const EditConnectionButton = ({
  connection,
  ...props
}: { connection: TDatadogConnection } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <EditConnectionModalContainer connection={connection} />

  return (
    <Button variant="ghost" onClick={() => addModal(modal)} {...props}>
      <Icon variant="PencilSimpleIcon" size={14} />
    </Button>
  )
}
