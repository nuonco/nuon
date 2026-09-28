import { useMemo } from 'react'
import { useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Link } from '@/components/common/Link'
import { BranchSettingsCards } from '@/components/branches/BranchSettingsPanel'
import { DeploymentPlanSection } from '@/components/branches/DeploymentPlanSection'
import { EditDeploymentPlanButton } from '@/components/branches/DeploymentPlanEditor'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { BranchProvider } from '@/providers/branch-provider'
import { getAppInstalls } from '@/lib'
import { latestBranchConfig } from '@/utils/branch-utils'
import type { TInstall } from '@/types'

const BranchSettingsContent = () => {
  const { org } = useOrg()
  const { app, labelColors } = useApp()
  const { branch, refresh } = useBranch()
  const params = useParams()
  const orgId = org.id!
  const appId = app.id!
  const branchId = params.branchId as string
  const hasInstallSyncing = !!org?.features?.['app-install-syncing']

  const currentConfig = useMemo(() => latestBranchConfig(branch), [branch])

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
    <div className="flex flex-col gap-10 p-4 md:p-6">
      <PageTitle
        segments={[
          branch?.name ? `${branch.name} settings` : 'Settings',
          app?.name,
        ]}
      />
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
            variant="primary"
            label="Edit plan"
            onSuccess={refresh}
          />
        }
      />
      {hasInstallSyncing ? (
        <Link
          href={`/${orgId}/apps/${appId}/branches/${branchId}/install-configs`}
        >
          Install configs
        </Link>
      ) : null}
      <BranchSettingsCards />
    </div>
  )
}

export const BranchSettingsTab = () => {
  const hasNewAppIA = useNewAppIA()
  const params = useParams()
  const branchId = params.branchId as string

  if (!hasNewAppIA) {
    return (
      <BranchProvider branchId={branchId}>
        <BranchSettingsContent />
      </BranchProvider>
    )
  }

  return <BranchSettingsContent />
}
