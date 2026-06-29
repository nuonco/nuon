import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import {
  createVCSConnectionWebhookSubscription,
  getVCSConnectionRepos,
  getVCSConnectionWebhookSubscription,
} from '@/lib'
import type { TAPIError, TVCSConnection } from '@/types'
import { ConnectionDetail } from './ConnectionDetail'

interface IConnectionDetailContainer {
  vcs_connection: TVCSConnection
}

export const ConnectionDetailContainer = ({ vcs_connection }: IConnectionDetailContainer) => {
  const { connectionId } = useParams<{ connectionId: string }>()
  const { org } = useOrg()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const {
    data: repos,
    error: reposError,
    isLoading: isLoadingRepos,
  } = useQuery({
    queryKey: ['vcs-connection-repos', org?.id, connectionId],
    queryFn: () => getVCSConnectionRepos({ orgId: org!.id, connectionId: connectionId! }),
    enabled: !!org?.id && !!connectionId,
  })

  const {
    data: webhookSubscription,
    isLoading: isLoadingSubscription,
    error: subscriptionError,
  } = useQuery({
    queryKey: ['vcs-connection-webhook-subscription', org?.id, connectionId],
    queryFn: () => getVCSConnectionWebhookSubscription({ orgId: org!.id, connectionId: connectionId! }),
    enabled: !!org?.id && !!connectionId,
    retry: false,
  })

  const subscriptionQueried = !isLoadingSubscription

  const { mutate: createSubscription, isPending: isCreatingSubscription } = useMutation({
    mutationFn: () =>
      createVCSConnectionWebhookSubscription({ orgId: org!.id, connectionId: connectionId! }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vcs-connection-webhook-subscription', org?.id, connectionId] })
      addToast(
        <Toast heading="Creating webhook subscription" theme="info">
          <Text>Webhook subscription is being created. This may take a moment.</Text>
        </Toast>
      )
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Webhook subscription failed" theme="error">
          <Text>{err?.error || 'Unable to create webhook subscription.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <ConnectionDetail
      vcs_connection={vcs_connection}
      repos={repos}
      reposError={reposError}
      isLoadingRepos={isLoadingRepos}
      webhookSubscription={webhookSubscription}
      subscriptionQueried={subscriptionQueried}
      onCreateSubscription={() => createSubscription()}
      isCreatingSubscription={isCreatingSubscription}
    />
  )
}
