import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { getAppSecretsConfig } from '@/lib'
import { AwaitAzureDetails } from './AwaitAzureDetails'
import type { IStackDetails } from '../types'

export const AwaitAzureDetailsContainer = ({
  stack,
  step,
  loading,
}: IStackDetails) => {
  const { install } = useInstall()
  const { org } = useOrg()
  const { appConfig } = useInstallAppConfig()

  const { data: secretsConfig } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-secrets-config', org?.id, install?.app_id],
    queryFn: () =>
      getAppSecretsConfig({
        orgId: org.id,
        appId: install.app_id,
      }),
    enabled: !!org?.id && !!install?.app_id,
  })

  // The install's current app config is only the scope of this stack version if
  // the version was rendered from it. A step viewed after the install repinned
  // to a newer config would otherwise be judged against a scope its template
  // never had.
  const renderedFromCurrentConfig =
    stack?.versions?.at(0)?.app_config_id === appConfig?.id

  return (
    <AwaitAzureDetails
      stack={stack}
      step={step}
      loading={loading}
      installId={install?.id ?? ''}
      azureLocation={install?.azure_account?.location}
      secrets={secretsConfig?.secrets}
      deploymentScope={
        renderedFromCurrentConfig
          ? appConfig?.stack?.deployment_scope
          : undefined
      }
    />
  )
}
