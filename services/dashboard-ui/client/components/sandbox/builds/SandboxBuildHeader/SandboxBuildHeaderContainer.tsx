import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSandboxBuild } from '@/hooks/use-sandbox-build'
import { SandboxBuildHeader } from './SandboxBuildHeader'

export const SandboxBuildHeaderContainer = () => {
  const { app } = useApp()
  const { org } = useOrg()
  const { build } = useSandboxBuild()

  return <SandboxBuildHeader app={app} build={build} orgId={org?.id} />
}
