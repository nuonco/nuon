import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { AwaitGCPDetails } from './AwaitGCPDetails'
import type { IStackDetails } from '../types'

interface IAwaitGCPDetailsContainer extends IStackDetails {
  spaceliftEnabled?: boolean
}

export const AwaitGCPDetailsContainer = ({
  stack,
  step,
  spaceliftEnabled,
  loading,
}: IAwaitGCPDetailsContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  return (
    <AwaitGCPDetails
      stack={stack}
      step={step}
      loading={loading}
      orgId={org.id}
      installId={install?.id}
      gcpProjectId={install?.gcp_account?.project_id}
      gcpRegion={install?.gcp_account?.region}
      spaceliftEnabled={spaceliftEnabled}
      tfProvider={!!org?.features?.['stack-tf-provider']}
    />
  )
}
