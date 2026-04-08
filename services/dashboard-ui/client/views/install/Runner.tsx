import { useQuery } from '@tanstack/react-query'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import {
  ProcessCard,
  ProcessCardSkeleton,
} from '@/components/runners/ProcessCard'
import { RunnerRecentActivity } from '@/components/runners/RunnerRecentActivity'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RunnerProvider } from '@/providers/runner-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { getRunnerSettings, getRunnerProcesses } from '@/lib'

const RunnerContent = ({
  runnerId,
  installId,
}: {
  runnerId: string
  installId: string
}) => {
  const { org } = useOrg()
  const { runner } = useRunner()

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
        runnerId,
        status: 'pending,active,offline,pending-shutdown',
        limit: 2,
      }),
    refetchInterval: 10000,
    enabled: !!org?.id && !!runnerId,
  })

  const processes = processResult?.data ?? []

  return (
    <>
      <Text variant="base" weight="strong">
        Processes
      </Text>

      {processesLoading ? (
        <div className="@container">
          <div className="grid grid-cols-1 @4xl:grid-cols-2 gap-6">
            <ProcessCardSkeleton />
            <ProcessCardSkeleton />
          </div>
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
        <div className="@container">
          <div className="grid grid-cols-1 @4xl:grid-cols-2 gap-6">
            {processes.map((process) => (
              <ProcessCard
                key={process.id}
                process={process}
                settings={settings}
                shouldPoll
              />
            ))}
          </div>
        </div>
      )}

      <Text variant="base" weight="strong">
        Recent jobs
      </Text>
      <RunnerRecentActivity
        shouldPoll
        jobDetailBasePath={`/${org?.id}/installs/${installId}/runner`}
      />
    </>
  )
}

export const Runner = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  if (!install?.runner_id) {
    return (
      <PageSection>
        <PageTitle title={`Install runner | ${install?.name}`} />
        <Breadcrumbs
          breadcrumbs={[
            { path: `/${org?.id}`, text: org?.name },
            { path: `/${org?.id}/installs`, text: 'Installs' },
            {
              path: `/${org?.id}/installs/${install?.id}`,
              text: install?.name,
            },
            {
              path: `/${org?.id}/installs/${install?.id}/runner`,
              text: 'Install runner',
            },
          ]}
        />
        <EmptyState
          emptyTitle="No runner"
          emptyMessage="This install does not have a runner yet."
          variant="diagram"
        />
      </PageSection>
    )
  }

  return (
    <RunnerProvider runnerId={install.runner_id} shouldPoll>
      <SurfacesProvider>
        <PageTitle title={`Install runner | ${install?.name}`} />
        <Breadcrumbs
          breadcrumbs={[
            { path: `/${org?.id}`, text: org?.name },
            { path: `/${org?.id}/installs`, text: 'Installs' },
            {
              path: `/${org?.id}/installs/${install?.id}`,
              text: install?.name,
            },
            {
              path: `/${org?.id}/installs/${install?.id}/runner`,
              text: 'Install runner',
            },
          ]}
        />
        <PageSection>
          <HeadingGroup>
            <Text variant="base" weight="strong">
              Install runner
            </Text>
            <ID>{install?.runner_id}</ID>
          </HeadingGroup>
          <RunnerContent runnerId={install.runner_id} installId={install.id} />
        </PageSection>
      </SurfacesProvider>
    </RunnerProvider>
  )
}
