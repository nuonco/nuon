import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import {
  createDatadogEventSubscription,
  getDatadogConnections,
} from '@/lib'
import type { TAPIError } from '@/types'
import {
  CreateEventSubscriptionModal,
  type CreateEventSubscriptionInput,
} from './CreateEventSubscription'

const CreateEventSubscriptionModalContainer = (
  props: Record<string, any>
) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const connectionsQuery = useQuery({
    queryKey: ['datadog-connections', org.id],
    queryFn: () => getDatadogConnections({ orgId: org.id }),
  })

  const { mutate, isPending, error } = useMutation({
    mutationFn: (input: CreateEventSubscriptionInput) =>
      createDatadogEventSubscription({
        orgId: org.id,
        body: {
          connection_id: input.connectionId,
          match: input.match,
          interests: input.interests,
          additional_tags: input.additionalTags,
          alert_type_override: input.alertTypeOverride || undefined,
          priority_override: input.priorityOverride || undefined,
        },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['datadog-event-subscriptions', org.id],
      })
      addToast(
        <Toast heading="Subscription created" theme="success">
          <Text>Matching events will start streaming into the connection.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      const heading =
        err?.status === 409
          ? 'Scope already subscribed'
          : 'Unable to create subscription'
      addToast(
        <Toast heading={heading} theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <CreateEventSubscriptionModal
      connections={connectionsQuery.data ?? []}
      isPending={isPending}
      error={error}
      onSubmit={mutate}
      {...props}
    />
  )
}

export const CreateEventSubscriptionButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <CreateEventSubscriptionModalContainer />

  return (
    <Button variant="primary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="PlusIcon" />
      Subscribe events
    </Button>
  )
}
