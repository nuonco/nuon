import { InstallComponentsList } from '@/components/install-components/InstallComponentsList'
import { PageTitle } from '@/components/navigation/PageTitle'
import { InstallWorkflowPanelController } from '@/components/workflows/InstallWorkflowPanel'
import { useInstall } from '@/hooks/use-install'

export const NewInstallComponents = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Components', install?.name]} />
      <InstallComponentsList />
      <InstallWorkflowPanelController />
    </>
  )
}
