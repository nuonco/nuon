import { useQuery } from '@tanstack/react-query'
import { BackToTop } from '@/components/common/BackToTop'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { RunnerDetailsCard, RunnerDetailsCardSkeleton } from '@/components/runners/RunnerDetailsCard'
import { RunnerHealthCard } from '@/components/runners/RunnerHealthCard'
import { RunnerRecentActivity } from '@/components/runners/RunnerRecentActivity'
import { ManagementDropdown } from '@/components/runners/management/ManagementDropdown'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RunnerProvider } from '@/providers/runner-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getRunnerSettings } from '@/lib'
import type { TRunnerGroup } from '@/types'

const CONTAINER_ID = 'install-runner-page'

const RunnerContent = ({ runnerId, installId }: { runnerId: string; installId: string }) => {
  const { org } = useOrg()

  const { data: settingsResult, isLoading: isLoadingSettings } = useQuery({
    queryKey: ['runner-settings', org?.id, runnerId],
    queryFn: () => getRunnerSettings({ orgId: org.id, runnerId }),
    enabled: !!org?.id && !!runnerId,
  })

  const settings = settingsResult

  return (
    <>
      <div className="flex gap-4 justify-between">
        <hgroup>
          <Text variant="base" weight="strong">
            Install runner
          </Text>
        </hgroup>
        {settings ? (
          <ManagementDropdown
            settings={settings}
            isInstallRunner
          />
        ) : null}
      </div>

      <div className="flex flex-col @min-4xl:flex-row gap-6">
        {isLoadingSettings ? (
          <RunnerDetailsCardSkeleton className="flex-initial" />
        ) : settings ? (
          <RunnerDetailsCard
            className="md:flex-initial"
            runnerGroup={settings as unknown as TRunnerGroup}
            shouldPoll
          />
        ) : (
          <Card className="flex-auto">
            <EmptyState
              emptyMessage="Runner details will display here once available."
              emptyTitle="No runner details"
              variant="table"
            />
          </Card>
        )}

        <RunnerHealthCard className="flex-auto" shouldPoll />
      </div>

      <div className="flex flex-col gap-6">
        <Text variant="base" weight="strong">
          Recent activity
        </Text>
        <RunnerRecentActivity shouldPoll jobDetailBasePath={`/${org?.id}/installs/${installId}/runner`} />
      </div>
    </>
  )
}

export const Runner = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  if (!install?.runner_id) {
    return (
      <PageSection isScrollable>
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
      <PageSection id={CONTAINER_ID} className="@container" isScrollable>
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
        <RunnerContent runnerId={install.runner_id} installId={install.id} />
        <BackToTop containerId={CONTAINER_ID} />
      </PageSection>
      </SurfacesProvider>
    </RunnerProvider>
  )
}
