import { useQuery } from '@tanstack/react-query'
import { BackToTop } from '@/components/common/BackToTop'
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
import { ProcessCard, ProcessCardSkeleton } from '@/components/runners/ProcessCard'
import { RunnerRecentActivity } from '@/components/runners/RunnerRecentActivity'
import { useOrg } from '@/hooks/use-org'
import { getRunnerSettings, getRunnerProcesses } from '@/lib'
import { RunnerProvider } from '@/providers/runner-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'

const CONTAINER_ID = 'org-build-runner-page'

const heading = (
  <HeadingGroup>
    <Text variant="h3" weight="strong" level={1}>
      Build runner
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

  const { data: processResult, isLoading: processesLoading } = useQuery({
    queryKey: ['runner-processes-active', org?.id, runnerId],
    queryFn: () =>
      getRunnerProcesses({
        orgId: org.id,
        runnerId: runnerId!,
        status: 'active,offline,pending-shutdown',
        limit: 2,
      }),
    refetchInterval: 10000,
    enabled: !!org?.id && !!runnerId,
  })

  const processes = processResult?.data ?? []

  const breadcrumbs = (
    <>
      <PageTitle title={`Build runner | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org?.name },
          { path: `/${org.id}/runner`, text: 'Build runner' },
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
      <PageLayout className="pb-6" id={CONTAINER_ID} isScrollable>
        {breadcrumbs}
        <PageHeader>
          {heading}
        </PageHeader>

        <PageContent>
          <PageSection>
            <Text variant="base" weight="strong">
              Processes
            </Text>
            {processesLoading ? (
              <div className="flex flex-wrap gap-6">
                <ProcessCardSkeleton />
                <ProcessCardSkeleton />
              </div>
            ) : processes.length === 0 ? (
              <Card>
                <EmptyState
                  emptyTitle="No active processes"
                  emptyMessage="No runner processes are currently active or offline."
                  variant="table"
                />
              </Card>
            ) : (
              <div className="flex flex-wrap gap-6">
                {processes.map((process) => (
                  <ProcessCard
                    key={process.id}
                    process={process}
                    settings={settings}
                    shouldPoll
                  />
                ))}
              </div>
            )}
          </PageSection>

          <PageSection>
            <Text variant="base" weight="strong">
              Recent jobs
            </Text>
            <RunnerRecentActivity shouldPoll jobDetailBasePath={`/${org?.id}/runner`} />
          </PageSection>
        </PageContent>
        <BackToTop containerId={CONTAINER_ID} />
      </PageLayout>
      </SurfacesProvider>
    </RunnerProvider>
  )
}
