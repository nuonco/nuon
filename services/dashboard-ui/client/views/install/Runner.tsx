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
import { ManagementDropdownContainer } from '@/components/runners/management/ManagementDropdown'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RunnerProvider } from '@/providers/runner-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getRunnerSettings, getRunnerProcesses } from '@/lib'

const PROCESS_STATUS_FILTER = 'pending,active,offline,pending-shutdown'

const RunnerContent = ({
  runnerId,
  installId,
}: {
  runnerId: string
  installId: string
}) => {
  const { org } = useOrg()

  const { data: settings } = useQuery({
    queryKey: ['runner-settings', org?.id, runnerId],
    queryFn: () => getRunnerSettings({ orgId: org.id, runnerId }),
    enabled: !!org?.id && !!runnerId,
  })

  const { data: mngResult, isLoading: mngLoading } = useQuery({
    queryKey: ['runner-processes-active', org?.id, runnerId, 'mng'],
    queryFn: () =>
      getRunnerProcesses({
        orgId: org.id,
        runnerId,
        type: 'mng',
        status: PROCESS_STATUS_FILTER,
        limit: 1,
      }),
    refetchInterval: 10000,
    enabled: !!org?.id && !!runnerId,
  })

  const { data: instanceResult, isLoading: instanceLoading } = useQuery({
    queryKey: ['runner-processes-active', org?.id, runnerId, 'install'],
    queryFn: () =>
      getRunnerProcesses({
        orgId: org.id,
        runnerId,
        type: 'install',
        status: PROCESS_STATUS_FILTER,
        limit: 1,
      }),
    refetchInterval: 10000,
    enabled: !!org?.id && !!runnerId,
  })

  const mngProcess = mngResult?.data?.[0]
  const instanceProcess = instanceResult?.data?.[0]
  const processes = [mngProcess, instanceProcess].filter(
    (p): p is NonNullable<typeof p> => !!p
  )
  const processesLoading = mngLoading || instanceLoading

  return (
    <>
      <div className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Install runner
          </Text>
          <ID>{runnerId}</ID>
        </HeadingGroup>
        {settings && (
          <ManagementDropdownContainer
            isInstallRunner
            settings={settings}
            hasMngProcess={!mngLoading ? !!mngProcess : undefined}
            hasInstanceProcess={!instanceLoading ? !!instanceProcess : undefined}
          />
        )}
      </div>

      <Text variant="base" weight="strong">
        Processes
      </Text>

      {processesLoading ? (
        <div className="@container">
          <div className="grid grid-cols-1 @5xl:grid-cols-2 gap-6 items-start">
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
      ) : processes.length === 1 ? (
        <ProcessCard
          process={processes[0]}
          settings={settings}
          shouldPoll
        />
      ) : (
        <div className="@container">
          <div className="grid grid-cols-1 @5xl:grid-cols-2 gap-6 items-start">
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
          <RunnerContent runnerId={install.runner_id} installId={install.id} />
        </PageSection>
      </SurfacesProvider>
    </RunnerProvider>
  )
}
