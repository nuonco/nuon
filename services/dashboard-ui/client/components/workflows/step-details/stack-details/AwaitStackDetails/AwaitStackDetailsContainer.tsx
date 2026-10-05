import { useInstall } from '@/hooks/use-install'
import { AwaitStackDetails } from './AwaitStackDetails'
import type { IStackDetails } from '../types'

export const AwaitStackDetailsContainer = (props: IStackDetails) => {
  const { install } = useInstall()
  const spaceliftEnabled = true
  return (
    <AwaitStackDetails
      runnerType={install?.app_runner_config?.app_runner_type}
      spaceliftEnabled={spaceliftEnabled}
      {...props}
    />
  )
}
