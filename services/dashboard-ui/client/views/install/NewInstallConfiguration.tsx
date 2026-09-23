import type { ReactNode } from 'react'
import {
  InstallConfigurationAppBranch,
  InstallConfigurationConfigFile,
  InstallConfigurationInputs,
  InstallConfigurationOverrides,
} from '@/components/installs/InstallConfiguration'
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
    <InstallConfigurationAppBranch />
  </ConfigurationPage>
)

export const NewInstallInputs = () => (
  <ConfigurationPage title="Inputs">
    <InstallConfigurationInputs />
  </ConfigurationPage>
)

export const NewInstallConfigFile = () => (
  <ConfigurationPage title="Config file">
    <InstallConfigurationConfigFile />
  </ConfigurationPage>
)

export const NewInstallOverrides = () => (
  <ConfigurationPage title="Overrides">
    <InstallConfigurationOverrides />
  </ConfigurationPage>
)
