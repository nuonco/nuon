import { useQuery } from '@tanstack/react-query'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RunnerDetailsCard } from '@/components/runners/RunnerDetailsCard'
import { RunnerHealthCard } from '@/components/runners/RunnerHealthCard'
import { RunnerRecentActivity } from '@/components/runners/RunnerRecentActivity'
import { ManagementDropdown } from '@/components/runners/management/ManagementDropdown'
import { useOrg } from '@/hooks/use-org'
import { getRunnerSettings } from '@/lib'
import { RunnerProvider } from '@/providers/runner-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'

const heading = (
  <HeadingGroup>
    <Text variant="h3" weight="strong" level={1}>
      Builds
    </Text>
    <Text theme="neutral">
      View your organizations build runner performance and activities.
    </Text>
  </HeadingGroup>
)

export const BuildRunner = () => {
  const { org } = useOrg()
  const runnerId = org?.runner_group?.runners?.[0]?.id

  const { data: settings } = useQuery({
    queryKey: ['runner-settings', org?.id, runnerId],
    queryFn: () => getRunnerSettings({ orgId: org.id, runnerId }),
    enabled: !!org?.id && !!runnerId,
  })

  const breadcrumbs = (
    <>
      <PageTitle title={`Builds | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org?.name },
          { path: `/${org.id}/runner`, text: 'Builds' },
        ]}
      />
    </>
  )

  if (!runnerId) {
    return (
      <PageLayout>
        {breadcrumbs}
        <PageHeader>{heading}</PageHeader>
        <PageContent>
          <Card>
            <EmptyState
              emptyTitle="No build runner"
              emptyMessage="No build runner is configured for this organization."
              variant="table"
            />
          </Card>
        </PageContent>
      </PageLayout>
    )
  }

  return (
    <RunnerProvider runnerId={runnerId} shouldPoll>
      <SurfacesProvider>
      <PageLayout isScrollable>
        {breadcrumbs}
        <PageHeader className="flex items-center justify-between">
          {heading}
          {settings && <ManagementDropdown settings={settings} />}
        </PageHeader>

        <PageContent>
          <PageSection className="flex-row gap-6">
            <RunnerDetailsCard
              className="flex-initial"
              runnerGroup={org.runner_group}
              shouldPoll
            />
            <RunnerHealthCard className="flex-auto" shouldPoll />
          </PageSection>

          <div className="flex gap-6">
            <PageSection>
              <Text variant="base" weight="strong">
                Recent activity
              </Text>
              <RunnerRecentActivity shouldPoll />
            </PageSection>
          </div>
        </PageContent>
      </PageLayout>
      </SurfacesProvider>
    </RunnerProvider>
  )
}
