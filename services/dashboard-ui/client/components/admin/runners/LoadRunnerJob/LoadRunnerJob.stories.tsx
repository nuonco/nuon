export default {
  title: 'Admin/LoadRunnerJob',
}

import { LoadRunnerJob } from './LoadRunnerJob'

export const Default = () => (
  <LoadRunnerJob
    job={{ id: 'job-1', status: 'finished', updated_at: new Date().toISOString() } as any}
    error={null}
    isLoading={false}
    title="Last shut-down job"
  />
)

export const Loading = () => (
  <LoadRunnerJob
    job={undefined}
    error={null}
    isLoading={true}
    title="Last shut-down job"
  />
)

export const NoJob = () => (
  <LoadRunnerJob
    job={undefined}
    error={null}
    isLoading={false}
    title="Last shut-down job"
  />
)
