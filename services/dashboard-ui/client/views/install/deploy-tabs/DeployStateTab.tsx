import { useParams } from 'react-router'
import { useOutletContext } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { TerraformWorkspaceCard } from '@/components/terraform-workspace/TerraformWorkspaceCard'
import { EmptyState } from '@/components/common/EmptyState'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { getInstallComponent } from '@/lib'
import type { TDeployOutletContext } from './types'

export const DeployStateTab = () => {
  const { componentId, installId } = useParams()
  const { component } = useOutletContext<TDeployOutletContext>()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: installComponent } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-component', org?.id, installId, componentId],
    queryFn: () =>
      getInstallComponent({
        orgId: org.id,
        installId: installId!,
        componentId: componentId!,
      }),
    enabled: !!org?.id && !!installId && !!componentId,
  })

  const workspaceId = installComponent?.terraform_workspace?.id

  return (
    <>
      <PageTitle segments={['Deploy state', install?.name]} />
      {!workspaceId ? (
        <EmptyState
          variant="table"
          emptyTitle="No state available"
          emptyMessage="No Terraform workspace found for this component."
        />
      ) : (
        <TerraformWorkspaceCard
          workspaceId={workspaceId}
          componentType={component?.type}
        />
      )}
    </>
  )
}
