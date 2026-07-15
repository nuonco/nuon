import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { GenerateStackDetails } from './GenerateStackDetails'

export const GenerateStackDetailsContainer = () => {
  const { appConfig, isLoading } = useInstallAppConfig()

  return (
    <GenerateStackDetails appConfig={appConfig} isLoading={isLoading} />
  )
}
