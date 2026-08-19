import { useRunner } from '@/hooks/use-runner'
import { RunnerStatusBanner } from './RunnerStatusBanner'

export const RunnerStatusBannerContainer = () => {
  const { runner } = useRunner()

  return (
    <RunnerStatusBanner warnings={runner?.warnings} status={runner?.status} />
  )
}
