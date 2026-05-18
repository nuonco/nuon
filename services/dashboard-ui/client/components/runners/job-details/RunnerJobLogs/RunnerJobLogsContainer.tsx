import { useRunnerJob } from '@/hooks/use-runner-job'
import { RunnerJobLogs } from './RunnerJobLogs'

export const RunnerJobLogsContainer = () => {
  const { job } = useRunnerJob()
  return <RunnerJobLogs logStreamId={job.log_stream_id} />
}
