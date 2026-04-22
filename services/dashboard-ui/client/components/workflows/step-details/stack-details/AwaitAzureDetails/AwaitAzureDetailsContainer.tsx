import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getAppSecretsConfig } from '@/lib'
import { AwaitAzureDetails } from './AwaitAzureDetails'
import type { IStackDetails } from '../types'

export const AwaitAzureDetailsContainer = ({ stack, step }: IStackDetails) => {
  const { install } = useInstall()
  const { org } = useOrg()

  const { data: secretsConfig } = useQuery({
    queryKey: ['app-secrets-config', org?.id, install?.app_id],
    queryFn: () =>
      getAppSecretsConfig({
        orgId: org.id,
        appId: install.app_id,
      }),
    enabled: !!org?.id && !!install?.app_id,
  })

  return (
    <AwaitAzureDetails
      stack={stack}
      step={step}
      installId={install?.id ?? ''}
      azureLocation={install?.azure_account?.location}
      secrets={secretsConfig?.secrets}
    />
  )
}
