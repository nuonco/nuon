import { useParams, useSearchParams } from 'react-router-dom'
import { usePolling } from '@/hooks/use-polling'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import {
  WorkflowTimeline,
  WorkflowTimelineSkeleton,
} from '@/components/workflows/WorkflowTimeline'
import { ShowDriftScan } from '@/components/workflows/filters/ShowDriftScans'
import { WorkflowTypeFilter } from '@/components/workflows/filters/WorkflowTypeFilter'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import type { TWorkflow } from '@/types'

const WorkflowsError = () => (
  <div className="w-full">
    <Text>Error fetching recent workflows activity</Text>
  </div>
)

const WorkflowsWrapper = ({
  installId,
  orgId,
  offset,
  showDrift,
  type,
}: {
  installId: string
  orgId: string
  offset: string
  showDrift: boolean
  type: string
}) => {
  const {
    data: workflows,
    error,
    isLoading,
    headers,
  } = usePolling<TWorkflow[]>({
    path: `/api/ctl-api/v1/installs/${installId}/workflows?offset=${offset}${type ? `&type=${type}` : ''}${showDrift ? '&planonly=true' : ''}`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  const pagination = {
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  if (error) {
    return <WorkflowsError />
  }

  if (isLoading && !workflows) {
    return <WorkflowTimelineSkeleton />
  }

  return (
    <WorkflowTimeline
      initWorkflows={workflows || []}
      pagination={pagination}
      ownerId={installId}
      ownerType="installs"
      shouldPoll
      planonly={showDrift}
      type={type}
    />
  )
}

export default function InstallWorkflows() {
  const { installId, orgId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()

  const offset = searchParams.get('offset') || '0'
  const type = searchParams.get('type') || ''
  const showDrift = searchParams.get('drifts') !== 'false'

  if (!installId || !orgId) {
    return null
  }

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name,
          },
          {
            path: `/${orgId}/installs/${installId}/workflows`,
            text: 'Workflows',
          },
        ]}
      />
      <div className="flex items-center gap-4 justify-between">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Workflows
          </Text>
        </HeadingGroup>

        <div className="flex items-center gap-4">
          <ShowDriftScan />
          <WorkflowTypeFilter />
        </div>
      </div>

      <WorkflowsWrapper
        installId={installId}
        orgId={orgId}
        offset={offset}
        showDrift={showDrift}
        type={type}
      />

      <BackToTop />
    </PageSection>
  )
}