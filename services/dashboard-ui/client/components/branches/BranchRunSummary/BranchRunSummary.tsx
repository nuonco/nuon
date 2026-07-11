import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import { RunDeploymentGraph } from '@/components/branches/RunDeploymentGraph'
import type {
  TComponentBuild,
  TInstallAppConfigVersion,
  TAppBranchRun,
  TInstallGroupRun,
  TComponentType,
} from '@/types'

interface IBranchRunSummary {
  branchRun?: TAppBranchRun
  builds: TComponentBuild[]
  installUpdates: TInstallAppConfigVersion[]
  installGroupRuns: TInstallGroupRun[]
  orgId: string
  appId: string
  branchId: string
  runStatus: string
}

const BuildsSection = ({
  builds,
  orgId,
  appId,
}: {
  builds: TComponentBuild[]
  orgId: string
  appId: string
}) => {
  if (builds.length === 0) return null

  return (
    <div className="flex flex-col gap-3">
      <Text variant="base" weight="strong">
        Builds ({builds.length})
      </Text>
      <div className="flex flex-col border rounded-lg divide-y overflow-hidden">
        {builds.map((build) => {
          const componentType = build.component_config_connection?.type as TComponentType | undefined
          const buildUrl = build.component_id && build.id
            ? `/${orgId}/apps/${appId}/components/${build.component_id}/builds/${build.id}`
            : undefined

          return (
            <div
              key={build.id}
              className="flex items-center justify-between gap-3 px-4 py-3"
            >
              <div className="flex items-center gap-3 min-w-0">
                {componentType && componentType !== 'unknown' ? (
                  <ComponentType
                    type={componentType}
                    displayVariant="icon-only"
                    colorVariant="color"
                    iconSize="16"
                  />
                ) : (
                  <Icon variant="PackageIcon" size={16} className="shrink-0 text-cool-grey-400" />
                )}
                <div className="flex flex-col gap-0.5 min-w-0">
                  <div className="flex items-center gap-2">
                    <Text variant="body" weight="strong" className="truncate">
                      {build.component_name || build.component_id}
                    </Text>
                    {componentType && componentType !== 'unknown' && (
                      <Text variant="subtext" theme="neutral">
                        {componentType.replace(/_/g, ' ')}
                      </Text>
                    )}
                  </div>
                  {build.id && (
                    <Text variant="subtext" theme="neutral" family="mono">
                      {build.id}
                    </Text>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-3 shrink-0">
                <Status status={build.status_v2?.status || 'unknown'} variant="badge" />
                {buildUrl && (
                  <Link href={buildUrl} className="text-xs">
                    View
                  </Link>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

const InstallsSection = ({
  installGroupRuns,
  orgId,
}: {
  installGroupRuns: TInstallGroupRun[]
  orgId: string
}) => {
  if (installGroupRuns.length === 0) return null

  return (
    <div className="flex flex-col gap-3">
      <Text variant="base" weight="strong">
        Updated installs
      </Text>
      <RunDeploymentGraph installGroupRuns={installGroupRuns} orgId={orgId} />
    </div>
  )
}

export const BranchRunSummary = ({
  branchRun,
  builds,
  installUpdates,
  installGroupRuns,
  orgId,
  appId,
  branchId,
  runStatus,
}: IBranchRunSummary) => {
  if (builds.length === 0 && installGroupRuns.length === 0) return null

  return (
    <div className="flex flex-col gap-6">
      <BuildsSection builds={builds} orgId={orgId} appId={appId} />
      <InstallsSection installGroupRuns={installGroupRuns} orgId={orgId} />
    </div>
  )
}
