import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { DeploymentPlanSection } from '@/components/branches/DeploymentPlanSection'
import { EditDeploymentPlanButton } from '@/components/branches/DeploymentPlanEditor'
import { getAppInstalls } from '@/lib'
import { latestBranchConfig } from '@/utils/branch-utils'
import type { TInstall } from '@/types'
import { BranchDetail } from '../BranchDetail'
import { BranchTabPage } from './BranchTabPage'

const BranchPlanContent = () => {
  const { org } = useOrg()
  const { app, labelColors } = useApp()
  const { branch, refresh } = useBranch()
  const orgId = org.id!
  const appId = app.id!

  const currentConfig = useMemo(() => latestBranchConfig(branch), [branch])
  const branchId = branch?.id

  const { data: appInstallsResult } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-installs', orgId, appId, branchId],
    queryFn: () =>
      getAppInstalls({ appId, orgId, app_branch_id: branchId, limit: 100 }),
    enabled: !!orgId && !!appId && !!branchId,
    refetchInterval: 10000,
  })

  const installsById = useMemo(
    () =>
      (appInstallsResult?.data ?? []).reduce<Record<string, TInstall>>(
        (acc, install) => {
          acc[install.id] = install
          return acc
        },
        {}
      ),
    [appInstallsResult]
  )

  return (
    <BranchTabPage
      tab="Deployment plan"
      tabPath="plan"
      heading="Deployment plan"
      subheading="Group installs and control the rollout order for this branch."
    >
      <DeploymentPlanSection
        config={currentConfig}
        installsById={installsById}
        orgId={orgId}
        labelColors={labelColors}
        createAction={
          <EditDeploymentPlanButton
            branch={branch}
            currentConfig={currentConfig}
            variant="secondary"
            label="Create deployment plan"
            onSuccess={refresh}
          />
        }
        editAction={
          <EditDeploymentPlanButton
            branch={branch}
            currentConfig={currentConfig}
            variant="ghost"
            label="Edit plan"
            onSuccess={refresh}
          />
        }
      />
    </BranchTabPage>
  )
}

export const BranchPlanTab = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchPlanContent /> : <BranchDetail />
}
