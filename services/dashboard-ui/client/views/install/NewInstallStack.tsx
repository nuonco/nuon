import { InstallStack } from '@/components/stacks/InstallStack'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'

export const NewInstallStack = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Stack', install?.name]} />
      <InstallStack />
    </>
  )
}
