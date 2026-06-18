import type { TVCSConnection, TVCSConnectionReposResponse, TVCSWebhookSubscription } from '@/types'
import { GitHubAccountSection } from './GitHubAccountSection'
import { RepositoriesSection } from './RepositoriesSection'
import { WebhookSubscriptionSection } from './WebhookSubscriptionSection'

interface IConnectionDetail {
  vcs_connection: TVCSConnection
  repos?: TVCSConnectionReposResponse
  reposError?: any
  isLoadingRepos?: boolean
  webhookSubscription?: TVCSWebhookSubscription
  subscriptionQueried?: boolean
  onCreateSubscription?: () => void
  isCreatingSubscription?: boolean
}

export const ConnectionDetail = ({
  vcs_connection,
  repos,
  reposError,
  isLoadingRepos = false,
  webhookSubscription,
  subscriptionQueried = false,
  onCreateSubscription,
  isCreatingSubscription,
}: IConnectionDetail) => (
  <>
    <GitHubAccountSection vcs_connection={vcs_connection} />
    {subscriptionQueried && (
      <WebhookSubscriptionSection
        webhookSubscription={webhookSubscription}
        onCreateSubscription={onCreateSubscription}
        isCreating={isCreatingSubscription}
      />
    )}
    <RepositoriesSection repos={repos} error={reposError} isLoading={isLoadingRepos} />
  </>
)
