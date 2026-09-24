import { useInstall } from '@/hooks/use-install'
import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'
import { AwaitStackDetails } from './AwaitStackDetails'
import type { IStackDetails } from '../types'

export const AwaitStackDetailsContainer = (props: IStackDetails) => {
  const { install } = useInstall()
  const spaceliftEnabled = useOrgFeatureFlag('spacelift-install-stacks')
  return (
    <AwaitStackDetails
      runnerType={install?.app_runner_config?.app_runner_type}
      spaceliftEnabled={spaceliftEnabled}
      {...props}
    />
  )
}
