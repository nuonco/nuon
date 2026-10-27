import { Navigate, useParams } from 'react-router'
import { PageTitle } from '@/components/navigation/PageTitle'
import { PullRequestPage } from '@/components/pull-requests/PullRequestPage'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { usePRPreviewUI } from '@/hooks/use-pr-preview-ui'

export const PullRequest = () => {
  const enabled = usePRPreviewUI()
  const { org } = useOrg()
  const { app } = useApp()
  const params = useParams()

  if (org && !enabled) {
    return (
      <>
        <PageTitle segments={['Pull request', app?.name]} />
        <Navigate
          to={`/${org.id}/apps/${app?.id}/branches/${params.branchId}`}
          replace
        />
      </>
    )
  }

  return <PullRequestPage />
}
