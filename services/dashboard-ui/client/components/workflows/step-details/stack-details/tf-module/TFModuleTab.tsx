import { useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { CreateServiceAccountTokenModalContainer } from '@/components/service-accounts/ServiceAccountToken'
import { useConfig } from '@/hooks/use-config'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOIDCTrustPolicies } from '@/hooks/use-oidc-trust-policies'
import { useStackServiceAccount } from '@/hooks/use-stack-service-account'
import { useSurfaces } from '@/hooks/use-surfaces'
import {
  OIDC_AUTH_ENABLED,
  OIDCAuthPane,
  StaticTokenAuthPane,
} from './AuthPanes'
import {
  buildInputsBlock,
  buildSecretExports,
  buildSecretsBlock,
  buildSecretVariablesBlock,
} from './snippet'

export interface IMainTfParts {
  installId: string
  providerBlock: string
  inputsBlock: string
  secretsBlock: string
  secretVariablesBlock: string
}

export interface ITFModuleTab {
  orgId: string
  installId?: string
  buildMainTf: (parts: IMainTfParts) => string
}

export const TFModuleTab = ({
  orgId,
  installId,
  buildMainTf,
}: ITFModuleTab) => {
  const queryClient = useQueryClient()
  const { addModal } = useSurfaces()
  const [authMethod, setAuthMethod] = useState<'token' | 'oidc'>('token')
  const { data: trustPolicies } = useOIDCTrustPolicies({
    enabled: OIDC_AUTH_ENABLED && authMethod === 'oidc',
  })
  const {
    data: serviceAccount,
    isLoading,
    isError,
  } = useStackServiceAccount({
    installId,
    orgId,
    enabled: true,
  })

  const openCreateToken = () =>
    addModal(
      <CreateServiceAccountTokenModalContainer
        accountId={serviceAccount?.account_id ?? ''}
        identity={serviceAccount?.email ?? 'this install stack'}
        defaultDuration="24h"
        tokenName={`stack-${installId}`}
        onCreated={() =>
          queryClient.invalidateQueries({
            queryKey: ['stack-service-account', installId],
          })
        }
      />
    )

  const liveUntil =
    serviceAccount?.has_live_token && serviceAccount.expires_at
      ? new Date(serviceAccount.expires_at).toLocaleString()
      : null

  const { appConfig } = useInstallAppConfig()

  const customerInputs = useMemo(() => {
    const declared = appConfig?.input?.inputs ?? []
    const grouped = (appConfig?.input?.input_groups ?? []).flatMap(
      (group) => group.app_inputs ?? []
    )
    const seen = new Set<string>()
    return [...declared, ...grouped].filter((input) => {
      if (!input.name || input.source !== 'customer') return false
      if (seen.has(input.name)) return false
      seen.add(input.name)
      return true
    })
  }, [appConfig?.input?.inputs, appConfig?.input?.input_groups])

  const customerSecrets = useMemo(
    () =>
      (appConfig?.secrets?.secrets ?? []).filter(
        (secret) => !!secret.name && !secret.auto_generate
      ),
    [appConfig?.secrets?.secrets]
  )

  const inputsBlock = useMemo(
    () => buildInputsBlock(customerInputs),
    [customerInputs]
  )
  const secretsBlock = useMemo(
    () => buildSecretsBlock(customerSecrets),
    [customerSecrets]
  )
  const secretVariablesBlock = useMemo(
    () => buildSecretVariablesBlock(customerSecrets),
    [customerSecrets]
  )
  const secretExports = useMemo(
    () => buildSecretExports(customerSecrets),
    [customerSecrets]
  )

  const config = useConfig()
  const runnerApiUrl =
    config.runnerApiUrl || serviceAccount?.runner_api_url || ''
  const providerBlock = runnerApiUrl
    ? `provider "stack" {\n  api_url = "${runnerApiUrl}"\n}`
    : 'provider "stack" {}'

  const mainTf = buildMainTf({
    installId: installId ?? '<install-id>',
    providerBlock,
    inputsBlock,
    secretsBlock,
    secretVariablesBlock,
  })

  const authCmd = `export NUON_API_TOKEN='<api-token>'`
  const applyCmd = `terraform init && terraform apply`

  return (
    <div className="flex flex-col gap-4 pt-4">
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          1. Create your Terraform configuration
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>main.tf</code>
            </Text>
            <ClickToCopyButton textToCopy={mainTf} />
          </span>
          <Code variant="preformated">{mainTf}</Code>
          {runnerApiUrl ? null : (
            <Text variant="subtext" theme="neutral">
              This control plane has not published its runner API URL, so the
              provider will default to production. Set{' '}
              <code>NUON_RUNNER_API_URL</code> on the dashboard and reload.
            </Text>
          )}
        </Card>
        {secretExports ? (
          <Card>
            <span className="flex justify-between items-center">
              <Text>Export the app&apos;s secret values.</Text>
              <ClickToCopyButton textToCopy={secretExports} />
            </span>
            <Code variant="preformated">{secretExports}</Code>
          </Card>
        ) : null}
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <span className="flex justify-between items-center gap-4">
          <Text variant="base" weight="strong">
            2. Authenticate
          </Text>
          {OIDC_AUTH_ENABLED ? (
            <ToggleButton<'token' | 'oidc'>
              value={authMethod}
              onChange={setAuthMethod}
              options={[
                { value: 'token', label: 'Static token' },
                { value: 'oidc', label: 'OIDC' },
              ]}
            />
          ) : null}
        </span>
        {OIDC_AUTH_ENABLED && authMethod === 'oidc' ? (
          <OIDCAuthPane
            installId={installId}
            audience={runnerApiUrl}
            policyNames={(trustPolicies ?? [])
              .map((policy) => policy.name ?? '')
              .filter(Boolean)}
          />
        ) : (
          <StaticTokenAuthPane
            authCmd={authCmd}
            isLoading={isLoading}
            isError={isError}
            liveUntil={liveUntil}
            canCreate={!!serviceAccount?.account_id}
            onCreateToken={openCreateToken}
          />
        )}
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          3. Apply
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Initialize the module and create the stack</Text>
            <ClickToCopyButton textToCopy={applyCmd} />
          </span>
          <Code variant="preformated">{applyCmd}</Code>
        </Card>
      </div>
    </div>
  )
}
