import { useSearchParams } from 'react-router'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { WorkflowTimeline } from '@/components/workflows/WorkflowTimeline'
import { ShowDriftScan } from '@/components/workflows/filters/ShowDriftScans'
import { WorkflowTypeFilter } from '@/components/workflows/filters/WorkflowTypeFilter'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Workflows = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()

  const type = searchParams.get('type') || ''
  const showDrifts = searchParams.get('drifts') !== 'false'

  return (
    <PageSection isScrollable>
      <PageTitle title={`Workflows | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          { path: `/${org?.id}/installs/${install?.id}/workflows`, text: 'Workflows' },
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

      <WorkflowTimeline
        installId={install?.id}
        shouldPoll
        planonly={showDrifts}
        type={type}
      />
    </PageSection>
  )
}
