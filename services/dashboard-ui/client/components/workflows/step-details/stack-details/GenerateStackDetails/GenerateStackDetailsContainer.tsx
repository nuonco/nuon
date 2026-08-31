import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import type { TInstallStackVersion } from '@/types'
import { GenerateStackDetails } from './GenerateStackDetails'

interface IGenerateStackDetailsContainer {
  stackVersion?: TInstallStackVersion
}

export const GenerateStackDetailsContainer = ({
  stackVersion,
}: IGenerateStackDetailsContainer) => {
  const { appConfig, isLoading } = useInstallAppConfig()

  return (
    <GenerateStackDetails
      appConfig={appConfig}
      isLoading={isLoading}
      stackVersion={stackVersion}
    />
  )
}
