export default {
  title: 'Runners/InstallRunner',
}

import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { RunnerRecentActivity } from '@/components/runners/RunnerRecentActivity/RunnerRecentActivity'
import { InstallRunner, InstallRunnerMissing } from './InstallRunner'

const actions = (
  <Button variant="secondary">Manage runner</Button>
)

const processCard = (name: string, status: string) => (
  <Card className="!p-4 !gap-2">
    <Text variant="base" weight="strong">
      {name}
    </Text>
    <Text variant="subtext" theme="neutral">
      {status}
    </Text>
  </Card>
)

const twoProcesses = (
  <div className="@container">
    <div className="grid grid-cols-1 @5xl:grid-cols-2 gap-6 items-start">
      {processCard('Install process', 'Active')}
      {processCard('Manager process', 'Active')}
    </div>
  </div>
)

const jobs = (
  <RunnerRecentActivity
    jobs={[
      {
        id: 'job-1',
        type: 'deploy',
        status: 'succeeded',
        created_at: '2026-09-22T14:00:00Z',
        group: 'deploy',
      },
      {
        id: 'job-2',
        type: 'action',
        status: 'running',
        created_at: '2026-09-22T13:40:00Z',
        group: 'action',
      },
    ] as never[]}
    isLoading={false}
    hasNext={false}
    offset={0}
  />
)

export const Default = () => (
  <InstallRunner
    runnerId="run7y9d0a78qni2edx83oyiym9"
    actions={actions}
    processes={twoProcesses}
    recentJobs={jobs}
  />
)

export const Embedded = () => (
  <InstallRunner
    variant="embedded"
    runnerId="run7y9d0a78qni2edx83oyiym9"
    actions={actions}
    processes={twoProcesses}
  />
)

export const EmbeddedWithWarning = () => (
  <InstallRunner
    variant="embedded"
    runnerId="run7y9d0a78qni2edx83oyiym9"
    actions={actions}
    statusBanner={
      <Banner theme="warn">Runner has not heartbeated in 5 minutes.</Banner>
    }
    processes={processCard('Install process', 'Offline')}
  />
)

export const EmbeddedDisabled = () => (
  <InstallRunner
    variant="embedded"
    runnerId="run7y9d0a78qni2edx83oyiym9"
    actions={actions}
    statusBanner={
      <Banner theme="neutral">
        This runner is disabled, so it runs no jobs and reports no health.
        Re-enable it in the install stack to start running jobs.
      </Banner>
    }
    processes={processCard('Install process', 'Disabled')}
  />
)

export const EmbeddedLoading = () => (
  <InstallRunner
    variant="embedded"
    runnerId="run7y9d0a78qni2edx83oyiym9"
    actions={actions}
    processesLoading
  />
)

export const EmbeddedNoProcesses = () => (
  <InstallRunner
    variant="embedded"
    runnerId="run7y9d0a78qni2edx83oyiym9"
    actions={actions}
  />
)

export const Missing = () => <InstallRunnerMissing />
