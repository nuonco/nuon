import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { deleteDatadogEventSubscription } from '@/lib'
import type { TAPIError, TDatadogEventSubscription } from '@/types'

const DeleteEventSubscriptionModalContainer = (
  props: { subscription: TDatadogEventSubscription } & Record<string, any>
) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending } = useMutation({
    mutationFn: () =>
      deleteDatadogEventSubscription({
        orgId: org.id,
        subscriptionId: props.subscription.id!,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['datadog-event-subscriptions', org.id],
      })
      addToast(<Toast heading="Subscription deleted" theme="success" />)
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to delete subscription" theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading="Delete event subscription"
      primaryActionTrigger={{
        children: isPending ? 'Deleting…' : 'Delete',
        disabled: isPending,
        onClick: () => mutate(),
        variant: 'danger',
      }}
      {...props}
    >
      <Text>
        Stop streaming events that match this scope into the Datadog
        connection. Existing events in DD's event stream are not removed.
      </Text>
    </Modal>
  )
}

export const DeleteEventSubscriptionButton = ({
  subscription,
  ...props
}: {
  subscription: TDatadogEventSubscription
} & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = (
    <DeleteEventSubscriptionModalContainer subscription={subscription} />
  )

  return (
    <Button variant="ghost" onClick={() => addModal(modal)} {...props}>
      <Icon variant="TrashIcon" size={14} />
    </Button>
  )
}
