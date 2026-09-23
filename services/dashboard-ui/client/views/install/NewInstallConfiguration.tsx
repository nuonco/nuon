import type { ReactNode } from 'react'
import { InstallAppBranch } from '@/components/installs/configuration/InstallAppBranch'
import { InstallConfigFile } from '@/components/installs/configuration/InstallConfigFile'
import { InstallConfigInputs } from '@/components/installs/configuration/InstallConfigInputs'
import { InstallConfigOverrides } from '@/components/installs/configuration/InstallConfigOverrides'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'

const ConfigurationPage = ({
  children,
  title,
}: {
  children: ReactNode
  title: string
}) => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={[title, install?.name]} />
      {children}
    </>
  )
}

export const NewInstallAppBranch = () => (
  <ConfigurationPage title="App branch">
    <InstallAppBranch />
  </ConfigurationPage>
)

export const NewInstallInputs = () => (
  <ConfigurationPage title="Inputs">
    <InstallConfigInputs />
  </ConfigurationPage>
)

export const NewInstallConfigFile = () => (
  <ConfigurationPage title="Config file">
    <InstallConfigFile />
  </ConfigurationPage>
)

export const NewInstallOverrides = () => (
  <ConfigurationPage title="Overrides">
    <InstallConfigOverrides />
  </ConfigurationPage>
)
