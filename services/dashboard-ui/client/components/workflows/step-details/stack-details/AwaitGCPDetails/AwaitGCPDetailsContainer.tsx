import { useInstall } from '@/hooks/use-install'
import { AwaitGCPDetails } from './AwaitGCPDetails'
import type { IStackDetails } from '../types'

interface IAwaitGCPDetailsContainer extends IStackDetails {
  spaceliftEnabled?: boolean
  tfProvider?: boolean
}

export const AwaitGCPDetailsContainer = ({
  stack,
  step,
  spaceliftEnabled,
  tfProvider,
  loading,
}: IAwaitGCPDetailsContainer) => {
  const { install } = useInstall()
  return (
    <AwaitGCPDetails
      stack={stack}
      step={step}
      loading={loading}
      installId={install?.id}
      gcpProjectId={install?.gcp_account?.project_id}
      spaceliftEnabled={spaceliftEnabled}
      tfProvider={tfProvider}
    />
  )
}
