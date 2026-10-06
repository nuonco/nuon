import { InstallImagesList } from '@/components/install-components/InstallImagesList'
import { PageTitle } from '@/components/navigation/PageTitle'
import { InstallWorkflowPanelController } from '@/components/workflows/InstallWorkflowPanel'
import { useInstall } from '@/hooks/use-install'

export const NewInstallImages = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Images', install?.name]} />
      <InstallImagesList />
      <InstallWorkflowPanelController />
    </>
  )
}
