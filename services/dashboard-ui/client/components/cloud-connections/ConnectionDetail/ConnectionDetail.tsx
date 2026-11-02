import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type {
  TAPIError,
  TCloudConnection,
  TCloudConnectionSummary,
} from '@/types'
import { ConnectionPolicies, failedRunbookStep } from '../ConnectionPolicies'
import { ConnectionStatus } from '../ConnectionStatus'
import { PrincipalSharing } from '../PrincipalSharing'

export type TConnectionTab = 'overview' | 'installs' | 'verification'

export const ConnectionMetadata = ({
  connection,
}: {
  connection?: TCloudConnection
}) => {
  const values = [
    [
      'Cloud',
      connection?.platform === 'gcp'
        ? 'Google Cloud'
        : connection?.platform === 'azure'
          ? 'Azure'
          : 'AWS',
    ],
    [
      connection?.platform === 'azure'
        ? 'Subscription'
        : connection?.platform === 'gcp'
          ? 'Project'
          : 'Account',
      connection?.target_id,
    ],
    [
      connection?.platform === 'azure'
        ? 'Application (client) ID'
        : connection?.platform === 'gcp'
          ? 'Service account'
          : 'IAM role',
      connection?.principal,
    ],
    ...(connection?.platform === 'azure'
      ? [
          ['Tenant', connection.azure_tenant_id],
          ['Registry', connection.azure_registry_login_server],
          ['Registry resource ID', connection.azure_registry_id],
        ]
      : [
          ...(connection?.default_region
            ? [['Location', connection.default_region]]
            : []),
          ...(connection?.platform === 'gcp'
            ? [['Project number', connection.gcp_project_number]]
            : []),
        ]),
  ]
  return (
    <section
      className="flex w-full min-w-0 flex-col gap-4"
      aria-label="Saved connection details"
    >
      <SectionHeader title="Saved connection details" />
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {values.map(([label, value]) => (
          <LabeledValue key={label} label={label} loading={!connection}>
            <Text
              family={label === 'Cloud' ? 'sans' : 'mono'}
              variant="subtext"
              className="break-all"
            >
              {value}
            </Text>
          </LabeledValue>
        ))}
      </div>
    </section>
  )
}

export const ConnectionDetail = ({
  connection,
  basePath,
  tab,
  isVerifying = false,
  verificationTimedOut = false,
  onVerify,
  onDelete,
  error,
  installs,
  repositoryAccess,
  sharedConnections = [],
}: {
  connection?: TCloudConnection
  basePath: string
  tab: TConnectionTab
  isVerifying?: boolean
  verificationTimedOut?: boolean
  onVerify: () => void
  onDelete: () => void
  error?: TAPIError | null
  installs?: ReactNode
  repositoryAccess?: ReactNode
  sharedConnections?: TCloudConnectionSummary[]
}) => (
  <DetailPage
    header={
      <DetailHeader
        backLink={false}
        loading={!connection && !error}
        title={
          <Text family="mono" weight="strong" className="break-all">
            {connection?.name || 'Cloud connection'}
          </Text>
        }
        description="Manage the saved cloud identity and generate cloud setup commands."
        status={
          <ConnectionStatus connection={connection} isVerifying={isVerifying} />
        }
        metadata={<ConnectionMetadata connection={connection} />}
        actions={
          connection ? (
            <>
              <Button disabled={isVerifying} onClick={onVerify}>
                {isVerifying ? 'Re-verifying' : 'Re-verify'}
              </Button>
              <Dropdown
                id="cloud-connection-actions"
                aria-label="Connection actions"
                buttonText={<Icon variant="DotsThreeIcon" size={20} />}
                hideIcon
                variant="ghost"
                alignment="right"
              >
                <Menu>
                  <Button variant="danger" onClick={onDelete}>
                    Delete connection
                  </Button>
                </Menu>
              </Dropdown>
            </>
          ) : null
        }
      />
    }
    tabNav={{
      basePath,
      tabs: [
        { path: '/', text: 'Overview' },
        {
          path: '/installs',
          text: (
            <span className="flex items-center gap-2">
              Installs{' '}
              <Badge size="xs" loading={!connection}>
                {connection?.used_by?.installs ?? 0}
              </Badge>
            </span>
          ),
        },
        { path: '/verification', text: 'Verification' },
      ],
    }}
  >
    <FormErrorBanner error={error} fallback="Cloud connection failed to load" />
    {verificationTimedOut && (
      <Banner theme="info">Still checking — refresh in a moment</Banner>
    )}
    {connection && tab === 'overview' && (
      <div className="flex flex-col gap-6">
        <PrincipalSharing connections={sharedConnections} />
        <Link href={`${basePath}/setup`}>View identity setup</Link>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <LabeledValue label="Created">
            <Time time={connection.created_at} format="long-datetime" />
          </LabeledValue>
          <LabeledValue label="Access">
            {connection.preset === 'stacks'
              ? 'Manages install stacks'
              : connection.preset === 'ecr' ||
                  connection.preset === 'gar' ||
                  connection.preset === 'acr'
                ? 'Image repository access'
                : 'Custom policy'}
          </LabeledValue>
          {['ecr', 'gar', 'acr'].includes(connection.preset) && (
            <LabeledValue label="Saved image configurations">
              {connection.used_by?.image_configs ?? 0}
            </LabeledValue>
          )}
        </div>
        <ConnectionPolicies
          connection={connection}
          repositoryAccess={repositoryAccess}
        />
      </div>
    )}
    {tab === 'installs' && installs}
    {connection && tab === 'verification' && (
      <div className="flex flex-col gap-6">
        <div className="grid grid-cols-2 gap-4">
          <LabeledValue label="Last verified">
            {connection.last_verified_at ? (
              <Time time={connection.last_verified_at} format="long-datetime" />
            ) : (
              'Never'
            )}
          </LabeledValue>
          <LabeledValue label="Status">
            <ConnectionStatus
              connection={connection}
              isVerifying={isVerifying}
            />
          </LabeledValue>
          <LabeledValue label="Result" className="col-span-2">
            {isVerifying
              ? connection.preset === 'acr'
                ? 'Checking Entra federation, foreign-subject rejection, and ACR token exchange.'
                : connection.preset === 'gar'
                  ? 'Checking federation trust, service account impersonation, and foreign-subject rejection.'
                  : connection.preset === 'ecr'
                    ? 'Checking role trust, foreign-subject rejection, and ECR token retrieval.'
                    : 'Checking the OIDC exchange and account access.'
              : connection.status_message || 'Verification has not run yet.'}
          </LabeledValue>
        </div>
        {connection.preset === 'ecr' && (
          <Text variant="subtext" theme="neutral">
            Verification checks role trust, foreign-subject rejection, and ECR
            token retrieval. Cloud IAM controls repository access; a component
            build proves access when pulling the image.
          </Text>
        )}
        {connection.preset === 'gar' && (
          <Text variant="subtext" theme="neutral">
            Verification checks GCP STS exchange, service account impersonation,
            and foreign-subject rejection, not repository access. Cloud IAM
            controls repository access; a component build proves access when
            pulling the image.
          </Text>
        )}
        {connection.preset === 'acr' && (
          <Text variant="subtext" theme="neutral">
            Verification checks Entra federation, foreign-subject rejection, and
            ACR token exchange, not repository access. Cloud IAM controls
            repository access; a component build proves access when pulling the
            image.
          </Text>
        )}
        {connection.status === 'error' &&
          !isVerifying &&
          !verificationTimedOut && (
            <Banner theme="error">
              <div className="flex flex-col items-start gap-2">
                <Text>{connection.status_message}</Text>
                <Link
                  href={`${basePath}/setup#runbook-step-${failedRunbookStep(connection)}`}
                >
                  View runbook step {failedRunbookStep(connection)}
                </Link>
              </div>
            </Banner>
          )}
      </div>
    )}
  </DetailPage>
)
