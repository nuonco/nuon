import { useInstall } from '@/hooks/use-install'
import { AwaitStackDetails, AwaitStackDetailsSkeleton } from './AwaitStackDetails'
import type { IStackDetails } from '../types'

export const AwaitStackDetailsContainer = (props: IStackDetails) => {
  const { install } = useInstall()
  return (
    <AwaitStackDetails
      runnerType={install?.app_runner_config?.app_runner_type}
      {...props}
    />
  )
}

export const AwaitStackDetailsSkeletonContainer = () => {
  const { install } = useInstall()
  return (
    <AwaitStackDetailsSkeleton
      runnerType={install?.app_runner_config?.app_runner_type}
    />
  )
}
