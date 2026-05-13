import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deleteSlackChannelSubscription } from '@/lib'
import type { TAPIError, TSlackChannelSubscription } from '@/types'
import { DeleteChannelSubscriptionModal } from './DeleteChannelSubscription'

const channelLabelFor = (sub: TSlackChannelSubscription): string => {
  if (sub.channel_name) return `#${sub.channel_name}`
  return sub.channel_id ?? ''
}

const DeleteChannelSubscriptionModalContainer = ({
  subscription,
  ...props
}: { subscription: TSlackChannelSubscription } & Record<string, any>) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: () =>
      deleteSlackChannelSubscription({
        orgId: org.id,
        subId: subscription.id ?? '',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['slack-channel-subscriptions', org.id],
      })
      addToast(
        <Toast heading="Subscription removed" theme="success">
          <Text>Lifecycle events will no longer post to this channel.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to remove subscription" theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <DeleteChannelSubscriptionModal
      channelLabel={channelLabelFor(subscription)}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const DeleteChannelSubscriptionButton = ({
  subscription,
  ...props
}: {
  subscription: TSlackChannelSubscription
} & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = (
    <DeleteChannelSubscriptionModalContainer subscription={subscription} />
  )

  return (
    <Button
      variant="ghost"
      className="!text-red-800 dark:!text-red-500"
      onClick={() => addModal(modal)}
      {...props}
    >
      <Icon variant="TrashIcon" />
      Remove
    </Button>
  )
}
