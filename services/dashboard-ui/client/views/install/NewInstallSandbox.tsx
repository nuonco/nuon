import { PageTitle } from '@/components/navigation/PageTitle'
import { InstallSandbox } from '@/components/sandbox/InstallSandbox'
import { useInstall } from '@/hooks/use-install'

export const NewInstallSandbox = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Sandbox', install?.name]} />
      <InstallSandbox />
    </>
  )
}
