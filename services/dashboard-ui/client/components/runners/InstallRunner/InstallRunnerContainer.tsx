import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ProcessCard } from '@/components/runners/ProcessCard'
import { RunnerRecentActivity } from '@/components/runners/RunnerRecentActivity'
import { RunnerStatusBanner } from '@/components/runners/RunnerStatusBanner'
import { ManagementDropdownContainer } from '@/components/runners/management/ManagementDropdown'
import {
  InstallRunner,
  InstallRunnerMissing,
  type TInstallRunnerVariant,
} from './InstallRunner'
import { RunnerProvider } from '@/providers/runner-provider'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getRunnerSettings, getRunnerProcesses } from '@/lib'

const InstallRunnerContent = ({
  installId,
  runnerId,
  variant,
}: {
  installId: string
  runnerId: string
  variant?: TInstallRunnerVariant
}) => {
  const { org } = useOrg()

  const { data: settings } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['runner-settings', org?.id, runnerId],
    queryFn: () => getRunnerSettings({ orgId: org.id, runnerId }),
    enabled: !!org?.id,
  })

  const { data: processResult, isLoading: processesLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['runner-processes-active', org?.id, runnerId],
    queryFn: () =>
      getRunnerProcesses({
        orgId: org.id,
        runnerId,
        status: 'pending,active,offline,pending-shutdown',
        limit: 2,
      }),
    refetchInterval: 10000,
    enabled: !!org?.id,
  })

  const processes = processResult?.data ?? []
  const runnerBasePath = `/${org?.id}/installs/${installId}/runner`

  let processesContent = null
  if (!processesLoading && processes.length === 1) {
    processesContent = (
      <ProcessCard
        process={processes[0]}
        settings={settings}
        shouldPoll
        runnerBasePath={runnerBasePath}
      />
    )
  } else if (!processesLoading && processes.length > 1) {
    processesContent = (
      <div className="@container">
        <div className="grid grid-cols-1 @5xl:grid-cols-2 gap-6 items-start">
          {processes.map((process) => (
            <ProcessCard
              key={process.id}
              process={process}
              settings={settings}
              shouldPoll
              runnerBasePath={runnerBasePath}
            />
          ))}
        </div>
      </div>
    )
  }

  return (
    <InstallRunner
      variant={variant}
      runnerId={runnerId}
      actions={
        settings ? (
          <ManagementDropdownContainer isInstallRunner settings={settings} />
        ) : null
      }
      statusBanner={<RunnerStatusBanner />}
      processesLoading={processesLoading}
      processes={processesContent}
      recentJobs={
        variant === 'embedded' ? null : (
          <RunnerRecentActivity shouldPoll jobDetailBasePath={runnerBasePath} />
        )
      }
    />
  )
}

export const InstallRunnerContainer = ({
  variant,
}: {
  variant?: TInstallRunnerVariant
} = {}) => {
  const { install } = useInstall()

  if (!install?.runner_id) {
    return <InstallRunnerMissing />
  }

  return (
    <RunnerProvider runnerId={install.runner_id} shouldPoll>
      <InstallRunnerContent
        variant={variant}
        runnerId={install.runner_id}
        installId={install.id}
      />
    </RunnerProvider>
  )
}
