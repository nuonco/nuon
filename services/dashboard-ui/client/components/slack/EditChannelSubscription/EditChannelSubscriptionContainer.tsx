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
import {
  EditChannelSubscriptionModal,
  type EditChannelSubscriptionInput,
} from './EditChannelSubscription'

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
    mutationFn: (input: EditChannelSubscriptionInput) =>
      updateSlackChannelSubscription({
        orgId: org.id,
        subId: subscription.id ?? '',
        body: {
          // PATCH treats `match` with PUT semantics — explicit `null`
          // resets to org-wide; an undefined wire value would mean
          // "leave unchanged". Map `undefined` to `null` so the modal
          // can clear the scope by toggling the radio back to
          // "Everything in this org".
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
    onError: (err: TAPIError) => {
      const heading =
        err?.status === 409
          ? 'Scope already subscribed to this channel'
          : 'Unable to save changes'
      addToast(
        <Toast heading={heading} theme="error">
          <Text>{err?.description || err?.error || 'Try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <EditChannelSubscriptionModal
      subscription={subscription}
      isPending={isPending}
      error={error as TAPIError | null}
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
      <Icon variant="PencilSimpleIcon" />
      Edit
    </Button>
  )
}
