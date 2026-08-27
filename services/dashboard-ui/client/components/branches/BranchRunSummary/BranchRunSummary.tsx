import { Text } from '@/components/common/Text'
import { BranchRunBuilds } from '@/components/branches/BranchRunBuilds'
import { RunDeploymentGraph } from '@/components/branches/RunDeploymentGraph'
import type {
  TComponentBuild,
  TInstall,
  TInstallAppConfigVersion,
  TAppBranchRun,
  TInstallGroupRun,
  TAppSandboxBuild,
} from '@/types'

interface IBranchRunSummary {
  branchRun?: TAppBranchRun
  builds: TComponentBuild[]
  sandboxBuild?: TAppSandboxBuild | null
  installUpdates: TInstallAppConfigVersion[]
  installGroupRuns: TInstallGroupRun[]
  installsById?: Record<string, TInstall>
  orgId: string
  appId: string
  branchId: string
  runStatus: string
}

const InstallsSection = ({
  installGroupRuns,
  installsById,
  orgId,
}: {
  installGroupRuns: TInstallGroupRun[]
  installsById?: Record<string, TInstall>
  orgId: string
}) => {
  if (installGroupRuns.length === 0) return null

  return (
    <div className="flex flex-col gap-3">
      <Text variant="base" weight="strong">
        Updated installs
      </Text>
      <RunDeploymentGraph
        installGroupRuns={installGroupRuns}
        installsById={installsById}
        orgId={orgId}
      />
    </div>
  )
}

export const BranchRunSummary = ({
  branchRun,
  builds,
  sandboxBuild,
  installUpdates,
  installGroupRuns,
  installsById,
  orgId,
  appId,
  branchId,
  runStatus,
}: IBranchRunSummary) => {
  if (builds.length === 0 && !sandboxBuild && installGroupRuns.length === 0) return null

  return (
    <div className="flex flex-col gap-6">
      <BranchRunBuilds
        builds={builds}
        sandboxBuild={sandboxBuild}
        orgId={orgId}
        appId={appId}
      />
      <InstallsSection
        installGroupRuns={installGroupRuns}
        installsById={installsById}
        orgId={orgId}
      />
    </div>
  )
}
