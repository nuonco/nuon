import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { deleteDatadogManagedMonitor } from '@/lib'
import type { TAPIError, TDatadogManagedMonitor } from '@/types'

const DeleteManagedMonitorModalContainer = (
  props: { monitor: TDatadogManagedMonitor } & Record<string, any>
) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending } = useMutation({
    mutationFn: () =>
      deleteDatadogManagedMonitor({
        orgId: org.id,
        monitorId: props.monitor.id!,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['datadog-managed-monitors', org.id],
      })
      addToast(
        <Toast heading="Monitor deleted" theme="success">
          <Text>The Datadog monitor has been removed.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to delete monitor" theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading="Delete managed monitor"
      primaryActionTrigger={{
        children: isPending ? 'Deleting…' : 'Delete monitor',
        disabled: isPending,
        onClick: () => mutate(),
        variant: 'danger',
      }}
      {...props}
    >
      <Text>
        The monitor is removed from Datadog. To recreate it, click "Alert in
        Datadog" again on the same Nuon resource.
      </Text>
    </Modal>
  )
}

export const DeleteManagedMonitorButton = ({
  monitor,
  ...props
}: {
  monitor: TDatadogManagedMonitor
} & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <DeleteManagedMonitorModalContainer monitor={monitor} />

  return (
    <Button variant="ghost" onClick={() => addModal(modal)} {...props}>
      <Icon variant="TrashIcon" size={14} />
    </Button>
  )
}
