import { InstallState } from '@/components/installs/InstallState'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'

export const NewInstallState = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['State', install?.name]} />
      <InstallState />
    </>
  )
}
