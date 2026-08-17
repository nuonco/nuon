import { Text } from '@/components/common/Text'
import { BranchRunBuilds } from '@/components/branches/BranchRunBuilds'
import { RunDeploymentGraph } from '@/components/branches/RunDeploymentGraph'
import type {
  TComponentBuild,
  TInstall,
  TInstallAppConfigVersion,
  TAppBranchRun,
  TInstallGroupRun,
} from '@/types'

interface IBranchRunSummary {
  branchRun?: TAppBranchRun
  builds: TComponentBuild[]
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
}: {
  installGroupRuns: TInstallGroupRun[]
  installsById?: Record<string, TInstall>
}) => {
  if (installGroupRuns.length === 0) return null

  return (
    <div className="flex flex-col gap-3">
      <Text variant="base" weight="strong">
        Updated installs
      </Text>
      <RunDeploymentGraph installGroupRuns={installGroupRuns} installsById={installsById} />
    </div>
  )
}

export const BranchRunSummary = ({
  branchRun,
  builds,
  installUpdates,
  installGroupRuns,
  installsById,
  orgId,
  appId,
  branchId,
  runStatus,
}: IBranchRunSummary) => {
  if (builds.length === 0 && installGroupRuns.length === 0) return null

  return (
    <div className="flex flex-col gap-6">
      <BranchRunBuilds builds={builds} orgId={orgId} appId={appId} />
      <InstallsSection installGroupRuns={installGroupRuns} installsById={installsById} />
    </div>
  )
}
