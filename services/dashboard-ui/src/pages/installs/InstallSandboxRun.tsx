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
import { Status } from '@/components/common/Status'
import { Time } from '@/components/common/Time'
import { Duration } from '@/components/common/Duration'
import type { TSandboxRun } from '@/types'

export default function InstallSandboxRun() {
  const { org } = useOrg()
  const { install } = useInstall()
  const { runId, orgId, installId } = useParams()

  const { data: sandboxRun, isLoading } = usePolling<TSandboxRun>({
    path: `/api/ctl-api/v1/installs/${installId}/sandbox/runs/${runId}`,
    pollInterval: 20000,
    shouldPoll: true,
  })

  if (isLoading) {
    return (
      <PageSection isScrollable>
        <Loading variant="stack" loadingText="Loading sandbox run details..." />
      </PageSection>
    )
  }

  if (!sandboxRun) {
    return (
      <PageSection isScrollable>
        <Text theme="neutral">Sandbox run not found.</Text>
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
          { path: `/${orgId}/installs/${installId}/sandbox`, text: 'Sandbox' },
          {
            path: `/${orgId}/installs/${installId}/sandbox/${runId}`,
            text: `Run ${runId}`,
          },
        ]}
      />
      <HeadingGroup>
        <div className="flex items-center gap-3">
          <Text variant="h3" weight="stronger" level={1}>
            Sandbox Run
          </Text>
          <Status status={sandboxRun?.status_v2?.status} variant="badge" />
        </div>
        <ID>{runId}</ID>
      </HeadingGroup>

      <div className="flex flex-col gap-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <Text variant="subtext" theme="neutral" className="mb-2">
              Status
            </Text>
            <Text>
              {sandboxRun?.status_v2?.status_human_description || 'Unknown'}
            </Text>
          </div>
          
          {sandboxRun?.created_at && (
            <div>
              <Text variant="subtext" theme="neutral" className="mb-2">
                Created
              </Text>
              <Time time={sandboxRun.created_at} />
            </div>
          )}

          {sandboxRun?.updated_at && (
            <div>
              <Text variant="subtext" theme="neutral" className="mb-2">
                Updated
              </Text>
              <Time time={sandboxRun.updated_at} />
            </div>
          )}

          {sandboxRun?.execution_time && (
            <div>
              <Text variant="subtext" theme="neutral" className="mb-2">
                Duration
              </Text>
              <Duration nanoseconds={sandboxRun.execution_time} />
            </div>
          )}
        </div>
      </div>

      <BackToTop />
    </PageSection>
  )
}
