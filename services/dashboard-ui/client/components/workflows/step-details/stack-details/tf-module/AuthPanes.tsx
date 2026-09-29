import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Text } from '@/components/common/Text'
import { CreateOIDCTrustPolicyButton } from '@/components/oidc-trust-policies'

export const OIDC_AUTH_ENABLED = false

export const StaticTokenAuthPane = ({
  authCmd,
  isLoading,
  isError,
  liveUntil,
  canCreate,
  onCreateToken,
}: {
  authCmd: string
  isLoading: boolean
  isError: boolean
  liveUntil: string | null
  canCreate: boolean
  onCreateToken: () => void
}) => {
  return (
    <>
      {isError ? (
        <Text variant="subtext" theme="neutral">
          This install stack has no service account yet. Reprovision the install
          to create one, then reload this page.
        </Text>
      ) : null}
      <Card>
        <span className="flex justify-between items-center">
          <Text>
            This token authorizes the module to read its configuration. Treat it
            as a secret.
          </Text>
          <ClickToCopyButton textToCopy={authCmd} />
        </span>
        <Code variant="preformated">{authCmd}</Code>
        {isLoading || isError || !canCreate ? null : (
          <span className="flex justify-between items-center gap-4">
            <Text variant="subtext" theme="neutral">
              {liveUntil
                ? `A token is active until ${liveUntil}. Creating another does not revoke it unless you ask it to.`
                : 'No token yet. Create one to get its value — it is shown only once.'}
            </Text>
            <Button size="sm" variant="secondary" onClick={onCreateToken}>
              Create token
            </Button>
          </span>
        )}
      </Card>
    </>
  )
}

export const OIDCAuthPane = ({
  installId,
  policyNames,
  audience,
}: {
  installId?: string
  policyNames: string[]
  audience: string
}) => {
  const policyName = `stack-${installId ?? 'install'}`
  const existing = policyNames.includes(policyName)

  const workflowSnippet = `permissions:
  id-token: write
  contents: read

env:
  NUON_ORG_ID: \${{ vars.NUON_ORG_ID }}`

  return (
    <>
      <Card>
        <span className="flex justify-between items-center gap-4">
          <Text>
            A trust policy tells Nuon which repository and branch may exchange
            an OIDC token for access to this org.
          </Text>
          {audience ? (
            <CreateOIDCTrustPolicyButton
              variant="secondary"
              size="sm"
              lockPreset
              repoSource="manual"
              githubAudience={audience}
              defaultRole="org_admin"
              defaultName={policyName}
              reservedNames={policyNames}
            >
              {existing ? 'Create another' : 'Create trust policy'}
            </CreateOIDCTrustPolicyButton>
          ) : null}
        </span>
        {audience ? null : (
          <Text variant="subtext" theme="neutral">
            This control plane has not published its runner API URL, which the
            policy needs as its audience. Set <code>NUON_RUNNER_API_URL</code>{' '}
            on the dashboard and reload.
          </Text>
        )}
        {existing ? (
          <Text variant="subtext" theme="neutral">
            A policy named <code>{policyName}</code> already exists. Check it
            covers the repository and branch running this Terraform before
            creating another.
          </Text>
        ) : null}
      </Card>

      <Card>
        <span className="flex justify-between items-center">
          <Text>
            Add this to the workflow that applies the Terraform. No{' '}
            <code>NUON_API_TOKEN</code>.
          </Text>
          <ClickToCopyButton textToCopy={workflowSnippet} />
        </span>
        <Code variant="preformated">{workflowSnippet}</Code>
      </Card>

      <Text variant="subtext" theme="neutral">
        Without <code>id-token: write</code> the workflow cannot mint an ID
        token and the provider falls back to looking for a static token.
      </Text>
    </>
  )
}
