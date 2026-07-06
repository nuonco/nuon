import { useMemo } from 'react'
import { Skeleton } from '@/components/common/Skeleton'
import { useBranch } from '@/hooks/use-branch'
import { useSurfaces } from '@/hooks/use-surfaces'
import { BranchProvider } from '@/providers/branch-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TAppBranch } from '@/types'
import { EditBranchButton } from '@/components/branches/EditBranchNameModal'
import { EditDeploymentPlanButton } from '@/components/branches/DeploymentPlanEditor'
import { TriggerBranchRunModal } from '@/components/branches/TriggerBranchRunModal'
import { BranchManagementDropdown } from './BranchManagementDropdown'

interface IBranchManagementMenu {
  appId: string
  orgId: string
}

const BranchManagementMenu = ({ appId, orgId }: IBranchManagementMenu) => {
  const { branch, refresh } = useBranch()
  const { addModal } = useSurfaces()

  const currentConfig = useMemo(() => {
    if (!branch.configs?.length) return undefined
    return [...branch.configs].sort(
      (a, b) => (b.config_number || 0) - (a.config_number || 0)
    )[0]
  }, [branch.configs])

  const handleTriggerRun = () => {
    addModal(
      <TriggerBranchRunModal
        branch={branch}
        currentConfig={currentConfig}
        appId={appId}
        orgId={orgId}
        planOnly={false}
        onSuccess={refresh}
      />
    )
  }

  return (
    <BranchManagementDropdown
      dropdownId={branch.id!}
      detailHref={`/${orgId}/apps/${appId}/branches/${branch.id}`}
      editButton={
        <EditBranchButton
          branch={branch}
          currentConfig={currentConfig}
          onSuccess={refresh}
          isMenuButton
        />
      }
      deploymentPlanButton={
        <EditDeploymentPlanButton
          branch={branch}
          currentConfig={currentConfig}
          onSuccess={refresh}
          isMenuButton
        />
      }
      hasConfig={!!currentConfig}
      isTriggerPending={false}
      onTriggerRun={handleTriggerRun}
    />
  )
}

export const BranchManagementDropdownContainer = ({
  branch,
  appId,
  orgId,
}: {
  branch: TAppBranch
  appId: string
  orgId: string
}) => {
  return (
    <BranchProvider
      branchId={branch.id!}
      shouldPoll={false}
      loadingElement={<Skeleton height="24px" width="24px" />}
      errorElement={null}
    >
      <SurfacesProvider>
        <BranchManagementMenu appId={appId} orgId={orgId} />
      </SurfacesProvider>
    </BranchProvider>
  )
}
