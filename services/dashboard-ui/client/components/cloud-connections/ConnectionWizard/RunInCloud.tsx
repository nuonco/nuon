import { useEffect, type ReactNode } from 'react'
import { useLocation } from 'react-router'
import { Banner } from '@/components/common/Banner'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { LabeledValue } from '@/components/common/LabeledValue'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TCloudConnection, TCloudConnectionSetup } from '@/types'
import {
  ConnectionPolicies,
  failedRunbookStep,
  roleName,
} from '../ConnectionPolicies'

export const RunInCloud = ({
  connection,
  setup,
  hasInitialSetup = false,
  repositoryAccess,
  setupMode,
}: {
  connection: TCloudConnection
  setup?: TCloudConnectionSetup
  hasInitialSetup?: boolean
  repositoryAccess?: ReactNode
  setupMode?: ReactNode
}) => {
  const { hash } = useLocation()
  const isGCP = connection.platform === 'gcp'
  const isAzure = connection.platform === 'azure'
  const isAWS = connection.platform === 'aws'
  const isImage = ['ecr', 'gar', 'acr'].includes(connection.preset)
  const hasInitialGrants = hasInitialSetup && !!setup?.permissions_policy
  useEffect(() => {
    if (hash)
      document.getElementById(hash.slice(1))?.scrollIntoView({ block: 'start' })
  }, [hash, setup])
  const snippet = (format: 'cli' | 'terraform' | 'cloudformation') =>
    (isAWS
      ? setup?.[format]?.replaceAll(
          '<role-name>',
          roleName(connection.principal)
        )
      : setup?.[format]) || ''
  return (
    <div className="flex min-w-0 flex-col gap-6">
      <SectionHeader
        title={
          isAzure
            ? 'Connect the identity in Azure'
            : isGCP
              ? 'Connect the identity in Google Cloud'
              : 'Connect the identity in AWS'
        }
        description={
          isAzure
            ? 'Complete the steps in order using an existing Entra application and service principal. Nuon does not provision Azure resources or store client secrets.'
            : isGCP
              ? 'Complete the steps in order. Nuon does not provision GCP resources or store service account keys.'
              : 'Complete the steps in order. Nuon never writes IAM resources into the account.'
        }
      />
      {connection.status === 'error' && (
        <Banner theme="error">
          Runbook step {failedRunbookStep(connection)} needs attention.{' '}
          {connection.status_message}
        </Banner>
      )}
      {isImage && (
        <Banner theme="info">
          {hasInitialGrants
            ? 'These setup commands include the repositories entered during creation. They are a setup snapshot, not an inventory of cloud grants.'
            : hasInitialSetup
              ? 'The connection details are saved. Run the identity setup below, then optionally generate repository access commands separately.'
              : 'This setup configures the cloud identity. Generate repository access commands separately; repository names are not saved to the connection.'}
        </Banner>
      )}
      {connection.preset === 'custom' && (
        <Banner theme="info">
          Custom access omits the permissions step. Attach your permissions
          policy to the role in AWS.
        </Banner>
      )}
      {isAzure && (
        <Banner theme="warn">
          <strong>Warning:</strong> ACR requires the RBAC + ABAC repository
          permissions mode. Do not switch an existing registry casually:
          changing modes can break existing role assignments and image access.
          Review the impact with the registry owner before applying setup.
        </Banner>
      )}
      <LabeledValue
        label={
          isAzure
            ? 'Application (client) ID'
            : isGCP
              ? 'Service account'
              : 'Role ARN'
        }
      >
        <ClickToCopy>
          <Text family="mono" variant="subtext" className="break-all">
            {connection.principal}
          </Text>
        </ClickToCopy>
      </LabeledValue>
      <ol className="flex list-decimal flex-col gap-2 pl-5">
        <li id="runbook-step-1" className="scroll-mt-6">
          <Text as="div" weight="strong">
            {isAzure
              ? 'Check the existing identity and registry prerequisites'
              : isGCP
                ? 'Configure the service account and connection-specific federation'
                : 'Create the OIDC provider'}
          </Text>
          <Text as="div" variant="subtext" theme="neutral">
            {isAzure
              ? 'Confirm the Entra application and its service principal exist in the selected tenant, the registry uses RBAC + ABAC repository permissions, and you can manage federated credentials and registry role assignments.'
              : isGCP
                ? 'Create or reuse the service account, then configure this connection’s provider with the Nuon issuer, audience, attribute mapping, and exact subject condition.'
                : 'Add the Nuon issuer to this AWS account once. If it already exists, reuse it.'}
          </Text>
        </li>
        <li>
          <Text as="div" weight="strong">
            {isAzure
              ? 'Add federation for this exact connection subject'
              : isGCP
                ? 'Grant service account impersonation to this connection'
                : 'Add this connection to the role trust policy'}
          </Text>
          <Text as="div" variant="subtext" theme="neutral">
            {isAzure
              ? 'Add a federated identity credential to the existing application with the exact Nuon issuer and subject and audience api://AzureADTokenExchange. The issuer discovery and signing keys must be publicly reachable by Microsoft Entra.'
              : isGCP
                ? 'Grant roles/iam.workloadIdentityUser on the service account to the exact subject rendered by Nuon, not the entire pool.'
                : 'Use the exact audience and connection subject rendered by Nuon. Preserve trust statements for other connections and services.'}
          </Text>
        </li>
        {(connection.preset === 'stacks' || (isImage && hasInitialGrants)) && (
          <li>
            <Text as="div" weight="strong">
              {isAzure
                ? 'Grant repository reader access with an ABAC condition'
                : isGCP
                  ? 'Grant reader access on each repository'
                  : 'Attach the permissions policy'}
            </Text>
            <Text as="div" variant="subtext" theme="neutral">
              {isAzure
                ? 'Assign Container Registry Repository Reader to the service principal at this registry, conditioned on the exact selected repository names. Do not use legacy AcrPull or grant registry catalog listing.'
                : isGCP
                  ? 'Grant roles/artifactregistry.reader to the service account on every selected repository, not on the project.'
                  : 'Review the IAM actions and scope below before attaching it.'}
            </Text>
          </li>
        )}
      </ol>
      {setupMode}
      <Expand
        id={`identity-setup-${connection.id}`}
        heading="View identity setup commands"
      >
        {!setup ? (
          <Text theme="neutral">Setup commands are not available yet.</Text>
        ) : (
          <Tabs
            naturalHeight
            initActiveTab="cli"
            tabLabels={{
              cli: isAzure ? 'Azure CLI' : isGCP ? 'gcloud CLI' : 'AWS CLI',
              terraform: 'Terraform',
              ...(isAWS ? { cloudformation: 'CloudFormation' } : {}),
            }}
            tabs={{
              cli: (
                <CodeBlock language="bash" showCopy wrapLongLines>
                  {snippet('cli')}
                </CodeBlock>
              ),
              terraform: (
                <CodeBlock language="hcl" showCopy>
                  {snippet('terraform')}
                </CodeBlock>
              ),
              ...(isAWS
                ? {
                    cloudformation: (
                      <div className="flex flex-col gap-3">
                        <Text variant="subtext" theme="neutral">
                          For an existing role, merge the trust contribution
                          into its owning stack and add only the permission
                          resources. Do not create another role resource.
                        </Text>
                        <CodeBlock language="yaml" showCopy>
                          {snippet('cloudformation')}
                        </CodeBlock>
                      </div>
                    ),
                  }
                : {}),
            }}
          />
        )}
      </Expand>
      <ConnectionPolicies
        connection={connection}
        initialSetup={hasInitialSetup ? setup : undefined}
        repositoryAccess={repositoryAccess}
      />
    </div>
  )
}
