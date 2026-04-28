import { useOrg } from '@/hooks/use-org'
import { DeployDetail } from './DeployDetail'
import { DeployLayout } from './DeployLayout'

export const DeployDetailGate = () => {
  const { org } = useOrg()

  if (org?.features?.['deploy-outputs']) {
    return <DeployLayout />
  }

  return <DeployDetail />
}
