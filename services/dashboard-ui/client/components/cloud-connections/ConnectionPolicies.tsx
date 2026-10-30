import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TCloudConnection, TCloudConnectionSetup } from '@/types'

export const roleName = (principal: string) =>
  principal.split(':role/')[1] || ''
export const awsRoleURL = (principal: string) =>
  `https://console.aws.amazon.com/iam/home#/roles/details/${encodeURIComponent(roleName(principal))}?section=trust`
export const identityConsoleURL = (connection: TCloudConnection) =>
  connection.platform === 'azure'
    ? `https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/ApplicationMenuBlade/~/Credentials/appId/${encodeURIComponent(connection.principal)}`
    : connection.platform === 'gcp'
      ? `https://console.cloud.google.com/iam-admin/serviceaccounts?project=${encodeURIComponent(connection.target_id)}`
      : awsRoleURL(connection.principal)
export const failedRunbookStep = (connection: TCloudConnection) =>
  connection.platform === 'azure'
    ? /ACR token exchange/i.test(connection.status_message || '')
      ? 3
      : /Entra|AADSTS|federat|trust|subject|audience|issuer|identity/i.test(
            connection.status_message || ''
          )
        ? 2
        : /ACR|repositor|registry|token|permission|ABAC/i.test(
              connection.status_message || ''
            )
          ? 3
          : 2
    : connection.platform === 'gcp'
      ? /identity|federation|trust|impersonat|\bsts\b/i.test(
          connection.status_message || ''
        )
        ? 2
        : /GAR|repositor|Artifact Registry/i.test(
              connection.status_message || ''
            )
          ? 3
          : 2
      : /Create the OIDC provider/i.test(connection.status_message || '')
        ? 1
        : (connection.preset === 'stacks' &&
              /CloudFormation/i.test(connection.status_message || '')) ||
            (connection.preset === 'ecr' &&
              /ECR/i.test(connection.status_message || ''))
          ? 3
          : 2

export const ConnectionPolicies = ({
  connection,
  initialSetup,
  repositoryAccess,
}: {
  connection: TCloudConnection
  initialSetup?: TCloudConnectionSetup
  repositoryAccess?: ReactNode
}) => {
  const setup = initialSetup || connection.setup
  const isGCP = connection.platform === 'gcp'
  const isAzure = connection.platform === 'azure'
  const isImage = ['ecr', 'gar', 'acr'].includes(connection.preset)
  return (
    <div className="flex min-w-0 flex-col gap-6">
      <section
        id="runbook-step-2"
        className="flex min-w-0 flex-col gap-3 scroll-mt-6"
      >
        <div className="flex items-center justify-between gap-3">
          <Text weight="strong">
            {isAzure
              ? 'Federated identity credential'
              : isGCP
                ? 'Federation configuration'
                : 'Trust policy'}
          </Text>
          <Link href={identityConsoleURL(connection)} isExternal>
            {isAzure
              ? 'View application in Azure portal'
              : isGCP
                ? 'View service accounts in GCP console'
                : 'View role in AWS console'}
          </Link>
        </div>
        <Text variant="subtext" theme="neutral">
          {isAzure
            ? 'Keep the exact issuer, connection subject, and api://AzureADTokenExchange audience unchanged. Microsoft Entra must be able to reach the issuer discovery document and signing keys publicly.'
            : isGCP
              ? 'Keep the issuer, audience, attribute mapping, and exact subject condition unchanged. Grant roles/iam.workloadIdentityUser only to this connection subject on the selected service account.'
              : 'Copy this policy as-is. To customize the role beyond it, edit it in the IAM console after creating it — the audience and subject conditions must stay unchanged for verification to pass.'}
        </Text>
        <Expand
          id={`connection-trust-${connection.id}`}
          heading="Preview identity trust configuration"
        >
          <CodeBlock language="json" showCopy wrapLongLines>
            {JSON.stringify(setup.trust_policy, null, 2)}
          </CodeBlock>
        </Expand>
      </section>
      {connection.preset === 'stacks' ||
      (isImage && initialSetup?.permissions_policy) ? (
        <section
          id={isImage ? undefined : 'runbook-step-3'}
          className="flex min-w-0 flex-col gap-3 scroll-mt-6"
        >
          <Text weight="strong">
            {isAzure
              ? 'Repository reader ABAC grant'
              : isGCP
                ? 'Repository role bindings'
                : 'Permissions policy'}
          </Text>
          {isAzure && (
            <Text variant="subtext" theme="neutral">
              Grant Container Registry Repository Reader at this registry with
              an ABAC condition for the exact selected repository names. This
              does not require legacy AcrPull or a catalog-listing role.
            </Text>
          )}
          <CodeBlock language="json" showCopy>
            {JSON.stringify(setup.permissions_policy, null, 2)}
          </CodeBlock>
        </section>
      ) : !isImage ? (
        <Banner theme="neutral">
          The permissions policy is managed in the customer's AWS account.
        </Banner>
      ) : null}
      {isImage && repositoryAccess}
    </div>
  )
}
