import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { DeploymentPlanGraph } from '@/components/branches/DeploymentPlanGraph'
import { InstallGroupsSection } from '@/components/branches/install-groups/InstallGroupsSection'
import type { TAppBranchConfig, TInstall } from '@/types'

interface IDeploymentPlanSection {
  config?: TAppBranchConfig
  installsById: Record<string, TInstall>
  orgId: string
  labelColors?: Record<string, string>
  createAction: ReactNode
  editAction?: ReactNode
}

export const DeploymentPlanSection = ({
  config,
  installsById,
  orgId,
  labelColors,
  createAction,
  editAction,
}: IDeploymentPlanSection) => {
  const hasDeploymentPlan = (config?.install_groups?.length ?? 0) > 0

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Text variant="base" weight="strong">
            Deployment plan
          </Text>
          {hasDeploymentPlan && (
            <Badge theme="info" size="sm">
              v{config?.config_number}
            </Badge>
          )}
        </div>
        {hasDeploymentPlan && editAction}
      </div>

      {hasDeploymentPlan && config ? (
        <div className="flex flex-col gap-6">
          <DeploymentPlanGraph
            config={config}
            installsById={installsById}
            orgId={orgId}
          />
          <div className="flex flex-col gap-3">
            <Text variant="subtext" weight="strong" theme="neutral">
              Install groups
            </Text>
            <InstallGroupsSection
              config={config}
              installsById={installsById}
              orgId={orgId}
              labelColors={labelColors}
            />
          </div>
        </div>
      ) : (
        <div className="border rounded-lg p-6">
          <EmptyState
            variant="diagram"
            emptyTitle="No deployment plan yet"
            emptyMessage="Create a deployment plan to group installs and roll out branch changes in stages."
            action={createAction}
          />
        </div>
      )}
    </div>
  )
}
