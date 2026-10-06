import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { getAppSecretsConfig, getInstallCurrentInputs } from '@/lib'
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

  const { data: currentInputs } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-current-inputs', org?.id, install?.id],
    queryFn: () =>
      getInstallCurrentInputs({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  // Names only. Sensitive inputs come back redacted, and knowing which inputs are
  // already set is enough to tell which the deploy command has to supply.
  const setInputNames = useMemo(
    () =>
      new Set(
        Object.entries(currentInputs?.values ?? {})
          .filter(([, value]) => !!value)
          .map(([name]) => name)
      ),
    [currentInputs?.values]
  )

  return (
    <AwaitAzureDetails
      stack={stack}
      step={step}
      loading={loading}
      orgId={org?.id ?? ''}
      installId={install?.id ?? ''}
      azureLocation={install?.azure_account?.location}
      azureSubscriptionId={install?.azure_account?.subscription_id}
      secrets={secretsConfig?.secrets}
      inputs={appConfig?.input?.inputs}
      setInputNames={setInputNames}
    />
  )
}
