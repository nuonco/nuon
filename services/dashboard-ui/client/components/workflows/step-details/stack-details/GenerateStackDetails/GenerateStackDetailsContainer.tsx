import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import type { TInstallStackVersionWithCompositeError } from '@/types'
import { GenerateStackDetails } from './GenerateStackDetails'

interface IGenerateStackDetailsContainer {
  stackVersion?: TInstallStackVersionWithCompositeError
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
