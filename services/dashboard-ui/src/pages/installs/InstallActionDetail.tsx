import { useParams } from 'react-router-dom'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { usePolling } from '@/hooks/use-polling'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { ID } from '@/components/common/ID'
import { Loading } from '@/components/common/Loading'
import { InstallActionsTable } from '@/components/actions/InstallActionsTable'
import type { TInstallAction } from '@/types'

const LIMIT = 10

export default function InstallActionDetail() {
  const { org } = useOrg()
  const { install } = useInstall()
  const { actionId, orgId, installId } = useParams()

  const { data: actionWithRuns, isLoading } = usePolling<TInstallAction>({
    path: `/api/ctl-api/v1/installs/${installId}/action-workflows/${actionId}/recent-runs?limit=${LIMIT}`,
    pollInterval: 30000,
    shouldPoll: true,
  })

  if (isLoading) {
    return (
      <PageSection isScrollable>
        <Loading variant="stack" loadingText="Loading action details..." />
      </PageSection>
    )
  }

  if (!actionWithRuns) {
    return (
      <PageSection isScrollable>
        <Text theme="neutral">Action not found.</Text>
      </PageSection>
    )
  }

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${orgId}`, text: org?.name || '' },
          { path: `/${orgId}/installs`, text: 'Installs' },
          { path: `/${orgId}/installs/${installId}`, text: install?.name || '' },
          { path: `/${orgId}/installs/${installId}/actions`, text: 'Actions' },
          {
            path: `/${orgId}/installs/${installId}/actions/${actionId}`,
            text: actionWithRuns?.action_workflow?.name || 'Action Detail',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="h3" weight="stronger" level={1}>
          {actionWithRuns?.action_workflow?.name}
        </Text>
        <ID>{actionId}</ID>
      </HeadingGroup>

      <div className="flex flex-col gap-6">
        <div>
          <Text variant="base" weight="strong" className="mb-4">
            Recent Runs
          </Text>
          <InstallActionsTable
            actionsWithRuns={actionWithRuns ? [actionWithRuns] : []}
            pagination={{
              limit: LIMIT,
              hasNext: false,
              offset: 0,
            }}
            shouldPoll={true}
          />
        </div>
      </div>

      <BackToTop />
    </PageSection>
  )
}
