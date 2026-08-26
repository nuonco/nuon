import { useOutletContext } from 'react-router'
import { TerraformWorkspaceCard } from '@/components/terraform-workspace/TerraformWorkspaceCard'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import type { TInstallComponentOutletContext } from './types'

export const InstallComponentStateTab = () => {
  const { install } = useInstall()
  const { installComponent } =
    useOutletContext<TInstallComponentOutletContext>()
  const component = installComponent?.component

  return (
    <>
      <PageTitle segments={['Component state', install?.name]} />
      <TerraformWorkspaceCard
        workspaceId={installComponent?.terraform_workspace?.id}
        componentType={component?.type}
      />
    </>
  )
}
