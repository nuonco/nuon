import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { updateSlackChannelSubscription } from '@/lib'
import type { TAPIError, TSlackChannelSubscription } from '@/types'
import { ChannelSubscriptionFormModal } from '@/components/slack/ChannelSubscriptionForm'
import type { ChannelSubscriptionOutput } from '@/components/slack/ChannelSubscriptionForm/schema'

const EditChannelSubscriptionModalContainer = ({
  subscription,
  ...props
}: {
  subscription: TSlackChannelSubscription
} & Record<string, any>) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: (input: ChannelSubscriptionOutput) =>
      updateSlackChannelSubscription({
        orgId: org.id,
        subId: subscription.id ?? '',
        body: {
          match: input.match ?? null,
          interests: input.interests,
        },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['slack-channel-subscriptions', org.id],
      })
      addToast(
        <Toast heading="Subscription updated" theme="success">
          <Text>Future events will use the new scope and event filter.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })

  const friendlyError =
    error && (error as TAPIError).status === 409
      ? {
          ...(error as TAPIError),
          error:
            'This channel is already subscribed with this scope. Pick a different scope.',
        }
      : (error as TAPIError | null)

  return (
    <ChannelSubscriptionFormModal
      mode="edit"
      subscription={subscription}
      isPending={isPending}
      error={friendlyError}
      onSubmit={mutate}
      {...props}
    />
  )
}

export const EditChannelSubscriptionButton = ({
  subscription,
  ...props
}: {
  subscription: TSlackChannelSubscription
} & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = (
    <EditChannelSubscriptionModalContainer subscription={subscription} />
  )

  return (
    <Button variant="ghost" onClick={() => addModal(modal)} {...props}>
      Edit subscription
      <Icon variant="PencilSimpleIcon" />
    </Button>
  )
}
