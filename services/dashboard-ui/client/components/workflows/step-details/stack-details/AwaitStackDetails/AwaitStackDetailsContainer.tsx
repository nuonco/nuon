import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { AwaitStackDetails } from './AwaitStackDetails'
import type { IStackDetails } from '../types'

export const AwaitStackDetailsContainer = (props: IStackDetails) => {
  const { install } = useInstall()
  const { org } = useOrg()
  return (
    <AwaitStackDetails
      runnerType={install?.app_runner_config?.app_runner_type}
      spaceliftEnabled={!!org?.features?.['spacelift-install-stacks']}
      tfProvider={!!org?.features?.['stack-tf-provider']}
      {...props}
    />
  )
}
