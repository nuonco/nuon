import { useRunnerJob } from '@/hooks/use-runner-job'
import { RunnerJobHeader } from './RunnerJobHeader'

export const RunnerJobHeaderContainer = () => {
  const { job } = useRunnerJob()
  return <RunnerJobHeader job={job} />
}
