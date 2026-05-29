import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { updateDatadogEventSubscription } from '@/lib'
import type {
  TAPIError,
  TDatadogAlertType,
  TDatadogConnection,
  TDatadogEventSubscription,
} from '@/types'
import {
  EditEventSubscriptionModal,
  type EditEventSubscriptionInput,
} from './EditEventSubscription'

const EditEventSubscriptionModalContainer = ({
  subscription,
  connection,
  ...props
}: {
  subscription: TDatadogEventSubscription
  connection: TDatadogConnection | undefined
} & Record<string, any>) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: (input: EditEventSubscriptionInput) =>
      updateDatadogEventSubscription({
        orgId: org.id,
        subscriptionId: subscription.id ?? '',
        body: {
          // PATCH treats `match` with PUT semantics — explicit null
          // resets to org-wide; undefined would mean "leave unchanged".
          match: input.match ?? null,
          interests: input.interests,
          additional_tags: input.additionalTags,
          // Empty string clears the override on the backend (the Go
          // validator uses `omitempty,oneof=...`, so "" passes and the
          // assignment writes the cleared value into the row).
          alert_type_override: (input.alertTypeOverride ||
            '') as TDatadogAlertType,
          priority_override: (input.priorityOverride || '') as any,
        },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['datadog-event-subscriptions', org.id],
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
          ? 'Scope already subscribed on this connection'
          : 'Unable to save changes'
      addToast(
        <Toast heading={heading} theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <EditEventSubscriptionModal
      subscription={subscription}
      connection={connection}
      isPending={isPending}
      error={error as TAPIError | null}
      onSubmit={mutate}
      {...props}
    />
  )
}

export const EditEventSubscriptionButton = ({
  subscription,
  connection,
  ...props
}: {
  subscription: TDatadogEventSubscription
  connection: TDatadogConnection | undefined
} & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = (
    <EditEventSubscriptionModalContainer
      subscription={subscription}
      connection={connection}
    />
  )

  return (
    <Button variant="ghost" onClick={() => addModal(modal)} {...props}>
      <Icon variant="PencilSimpleIcon" />
      Edit
    </Button>
  )
}
