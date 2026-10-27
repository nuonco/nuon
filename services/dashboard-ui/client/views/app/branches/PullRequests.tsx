import { Navigate, useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { PullRequestList } from '@/components/pull-requests/PullRequestList'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { usePRPreviewUI } from '@/hooks/use-pr-preview-ui'
import { getPreviewSources } from '@/lib'
import type { TAPIError } from '@/types'
import { ProviderError } from '@/components/layout/ProviderError'
import { PageTitle } from '@/components/navigation/PageTitle'
import { BranchTabPage } from './tabs/BranchTabPage'

export const PullRequests = () => {
  const enabled = usePRPreviewUI()
  const { org } = useOrg()
  const { app } = useApp()
  const params = useParams()
  const orgId = org?.id ?? ''
  const appId = app?.id ?? ''
  const branchId = params.branchId ?? ''

  const sources = useQuery({
    queryKey: ['preview-sources', orgId, appId, branchId],
    queryFn: () => getPreviewSources({ orgId, appId, branchId }),
    enabled: enabled && !!orgId && !!appId && !!branchId,
    placeholderData: keepPreviousData,
  })

  if (org && !enabled) {
    return (
      <>
        <PageTitle segments={['Pull requests', app?.name]} />
        <Navigate
          to={`/${org.id}/apps/${app?.id}/branches/${branchId}`}
          replace
        />
      </>
    )
  }

  return (
    <BranchTabPage
      tab="Pull requests"
      tabPath="pull-requests"
      heading="Pull requests"
      subheading="Open pull requests targeting this branch."
    >
      {sources.error && !sources.data ? (
        <ProviderError error={sources.error as TAPIError} />
      ) : (
        <PullRequestList
          pullRequests={sources.data?.pull_requests ?? []}
          basePath={`/${orgId}/apps/${appId}/branches/${branchId}`}
          isLoading={sources.isPending}
        />
      )}
    </BranchTabPage>
  )
}
