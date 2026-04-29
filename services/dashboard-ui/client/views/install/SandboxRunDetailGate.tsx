import { useOrg } from '@/hooks/use-org'
import { SandboxRunDetail } from './SandboxRunDetail'
import { SandboxRunLayout } from './SandboxRunLayout'

export const SandboxRunDetailGate = () => {
  const { org } = useOrg()

  if (org?.features?.['deploy-outputs']) {
    return <SandboxRunLayout />
  }

  return <SandboxRunDetail />
}
