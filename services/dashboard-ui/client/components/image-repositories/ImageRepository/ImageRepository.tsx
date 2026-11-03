import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import {
  CloudPlatform,
  cloudPlatformName,
} from '@/components/common/CloudPlatform'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Dropdown } from '@/components/common/Dropdown'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { roleName } from '@/components/cloud-connections/ConnectionPolicies'
import { ConnectionSetupMode } from '@/components/cloud-connections/ConnectionSetupMode'
import { PrincipalSharing } from '@/components/cloud-connections/PrincipalSharing'
import type {
  TAPIError,
  TCloudConnection,
  TCloudConnectionSetup,
  TCloudConnectionSummary,
  TImageRepository,
} from '@/types'

export const ImageRepository = ({
  orgId,
  repository,
  connection,
  setup,
  error,
  setupError,
  isVerifying = false,
  verificationTimedOut = false,
  onVerify,
  onReloadSetup,
  onRemove,
  sharedConnections = [],
  existingPrincipal = false,
  onSetupModeChange,
  verifyError,
}: {
  orgId: string
  repository?: TImageRepository
  connection?: TCloudConnection
  setup?: TCloudConnectionSetup
  error?: TAPIError | null
  setupError?: TAPIError | null
  isVerifying?: boolean
  verificationTimedOut?: boolean
  onVerify: () => void
  onReloadSetup: () => void
  onRemove?: () => void
  sharedConnections?: TCloudConnectionSummary[]
  existingPrincipal?: boolean
  onSetupModeChange?: (existing: boolean) => void
  verifyError?: TAPIError | null
}) => {
  const cloud = cloudPlatformName(repository?.platform)
  const verified = connection?.status === 'verified'
  const cli = setup?.cli.replaceAll(
    '<role-name>',
    roleName(connection?.principal || '')
  )
  const permissions = setup?.permissions_cli?.replaceAll(
    '<role-name>',
    roleName(connection?.principal || '')
  )
  const config = repository
    ? repository.platform === 'aws'
      ? `[aws_ecr]\nimage_url = "${repository.image_url_prefix}"\ntag = "latest"\nregion = "${repository.cloud_connection.default_region}"`
      : repository.platform === 'gcp'
        ? `[gcp_gar]\nimage_url = "${repository.image_url_prefix}/example-image"\ntag = "latest"\ngcp_project_id = "${repository.cloud_connection.target_id}"\nregion = "${repository.cloud_connection.default_region}"`
        : `[azure_acr]\nimage_url = "${repository.image_url_prefix}"\ntag = "latest"\nregistry_url = "${repository.cloud_connection.azure_registry_login_server}"`
    : ''
  return (
    <DetailPage
      header={
        <DetailHeader
          backLink={false}
          loading={!repository && !error}
          title={repository?.image_url_prefix || 'Image repository'}
          description="Set up image pull access, then use this repository in app configuration."
          metadata={
            repository ? (
              <>
                <LabeledValue label="Identity" loading={!connection}>
                  {repository.cloud_connection?.name ? (
                    <Text family="mono" className="break-all">
                      {repository.cloud_connection.name}
                    </Text>
                  ) : (
                    'Unavailable identity'
                  )}
                </LabeledValue>
                <LabeledStatus
                  label="Authentication"
                  loading={!connection}
                  statusProps={{
                    status: isVerifying ? 'pending' : connection?.status,
                  }}
                />
                <LabeledValue label="Cloud" loading={!connection}>
                  <CloudPlatform
                    platform={repository.platform}
                    displayVariant="name"
                    variant="subtext"
                  />
                </LabeledValue>
                <LabeledValue label="Principal" loading={!connection}>
                  <Text family="mono" variant="subtext" className="break-all">
                    {repository.cloud_connection?.principal}
                  </Text>
                </LabeledValue>
              </>
            ) : null
          }
          actions={
            repository && onRemove ? (
              <Dropdown
                id="image-repository-actions"
                aria-label="Repository actions"
                buttonText={<Icon variant="DotsThreeIcon" size={20} />}
                hideIcon
                variant="ghost"
                alignment="right"
              >
                <Menu>
                  <Button variant="danger" onClick={onRemove}>
                    Remove repository
                  </Button>
                </Menu>
              </Dropdown>
            ) : null
          }
        />
      }
    >
      <FormErrorBanner
        error={error}
        fallback="Image repository failed to load"
      />
      {repository && (
        <>
          <section className="flex flex-col gap-4" aria-label="Pull identity">
            <SectionHeader
              title="Pull identity"
              description="This selection is saved for this repository. Cloud permissions are configured separately."
            />
            <Link
              href={`/${orgId}/settings/cloud-connections/${repository.cloud_connection_id}`}
            >
              View cloud connection
            </Link>
            <PrincipalSharing connections={sharedConnections} orgId={orgId} />
          </section>
          <section
            className="flex min-w-0 flex-col gap-4 border-t pt-6"
            aria-label="Repository setup"
          >
            <SectionHeader
              title="Configure pull access"
              description={`Review and apply these commands in ${cloud}. Nuon does not apply cloud permissions for you.`}
            />
            {onSetupModeChange && (
              <ConnectionSetupMode
                platform={repository.platform}
                existingPrincipal={existingPrincipal}
                onChange={onSetupModeChange}
              />
            )}
            <FormErrorBanner
              error={setupError}
              fallback="Setup commands failed to load"
            />
            {setupError && (
              <Button className="w-fit" onClick={onReloadSetup}>
                Retry setup commands
              </Button>
            )}
            {repository.platform === 'azure' && (
              <Banner theme="warn">
                The registry must already use RBAC + ABAC repository
                permissions. Review existing role assignments before changing
                registry modes.
              </Banner>
            )}
            {setup && (
              <>
                <Expand
                  id="repository-identity-setup"
                  heading="Set up authentication"
                  isOpen={!verified}
                >
                  <div className="flex flex-col gap-4">
                    <Text variant="subtext" theme="neutral">
                      Configure federation for this exact connection, even when
                      its principal is shared. The CLI preserves existing trust
                      and permissions. For a previously configured connection,
                      use the permissions-only commands below.
                    </Text>
                    <Tabs
                      naturalHeight
                      initActiveTab="cli"
                      tabLabels={{
                        cli: `${cloud === 'Google Cloud' ? 'gcloud' : cloud} CLI`,
                        terraform: 'Terraform',
                        ...(setup.cloudformation
                          ? { cloudformation: 'CloudFormation' }
                          : {}),
                      }}
                      tabs={{
                        cli: (
                          <CodeBlock language="bash" showCopy wrapLongLines>
                            {cli || ''}
                          </CodeBlock>
                        ),
                        terraform: (
                          <CodeBlock language="hcl" showCopy>
                            {setup.terraform.replaceAll(
                              '<role-name>',
                              roleName(connection?.principal || '')
                            )}
                          </CodeBlock>
                        ),
                        ...(setup.cloudformation
                          ? {
                              cloudformation: (
                                <CodeBlock language="yaml" showCopy>
                                  {setup.cloudformation.replaceAll(
                                    '<role-name>',
                                    roleName(connection?.principal || '')
                                  )}
                                </CodeBlock>
                              ),
                            }
                          : {}),
                      }}
                    />
                  </div>
                </Expand>
                <Expand
                  id="repository-permissions"
                  heading="Grant repository permissions"
                  isOpen={verified}
                >
                  <div className="flex flex-col gap-4">
                    <Text variant="subtext" theme="neutral">
                      These additive commands grant this repository to the
                      saved principal without recreating authentication. If
                      cloud IaC manages its policy, update that source of truth
                      too.
                    </Text>
                    <CodeBlock language="bash" showCopy wrapLongLines>
                      {permissions || ''}
                    </CodeBlock>
                    {setup.permissions_policy && (
                      <Expand
                        id="repository-policy"
                        heading="Preview permissions policy"
                      >
                        <CodeBlock language="json" showCopy>
                          {JSON.stringify(setup.permissions_policy, null, 2)}
                        </CodeBlock>
                      </Expand>
                    )}
                  </div>
                </Expand>
              </>
            )}
          </section>
          <section
            className="flex flex-col gap-4 border-t pt-6"
            aria-label="Verify authentication"
          >
            <SectionHeader
              title="Verify authentication"
              description="This checks identity federation, not repository permissions. After verification, sync the app and build an image component to check pull access."
            />
            {connection?.status === 'error' && (
              <Banner theme="error">
                {connection.status_message ||
                  'Authentication failed. Review setup and try again.'}
              </Banner>
            )}
            {verificationTimedOut && (
              <Banner theme="info">Still checking — refresh in a moment</Banner>
            )}
            {verified && (
              <Banner theme="success">
                Authentication verified. A component build checks actual image
                access.
              </Banner>
            )}
            <FormErrorBanner
              error={verifyError}
              fallback="Verification failed to start"
            />
            <Button
              className="w-fit"
              variant="primary"
              disabled={!connection || isVerifying}
              onClick={onVerify}
            >
              {isVerifying
                ? 'Verifying authentication'
                : verified
                  ? 'Re-verify authentication'
                  : 'Verify authentication'}
            </Button>
          </section>
          <section
            className="flex min-w-0 flex-col gap-4 border-t pt-6"
            aria-label="App configuration"
          >
            <SectionHeader
              title="Use in app configuration"
              description={
                repository.platform === 'gcp'
                  ? 'Replace example-image and latest with the image path and tag. The repository mapping selects the pull identity; no connection ID is needed in config.'
                  : 'Set the intended tag. The repository mapping selects the pull identity; no connection ID is needed in config.'
              }
            />
            <CodeBlock language="toml" showCopy>
              {config}
            </CodeBlock>
          </section>
        </>
      )}
    </DetailPage>
  )
}
