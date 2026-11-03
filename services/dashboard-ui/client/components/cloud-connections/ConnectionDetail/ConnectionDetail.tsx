import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { CloudPlatform } from '@/components/common/CloudPlatform'
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
import type {
  TAPIError,
  TCloudConnection,
  TCloudConnectionSummary,
} from '@/types'
import { ConnectionPolicies, failedRunbookStep } from '../ConnectionPolicies'
import { ConnectionStatus } from '../ConnectionStatus'
import { PrincipalSharing } from '../PrincipalSharing'
import { isImagePreset } from '../repository-scope'

export type TConnectionTab = 'overview' | 'installs' | 'verification'

const verifyingCopy: Record<'ecr' | 'gar' | 'acr', string> = {
  acr: 'Checking Entra federation, foreign-subject rejection, and ACR token exchange.',
  gar: 'Checking federation trust, service account impersonation, and foreign-subject rejection.',
  ecr: 'Checking role trust, foreign-subject rejection, and ECR token retrieval.',
}
const imageNoteCopy: Record<'ecr' | 'gar' | 'acr', string> = {
  ecr: 'Verification checks role trust, foreign-subject rejection, and ECR token retrieval. Cloud IAM controls repository access; a component build proves access when pulling the image.',
  gar: 'Verification checks GCP STS exchange, service account impersonation, and foreign-subject rejection, not repository access. Cloud IAM controls repository access; a component build proves access when pulling the image.',
  acr: 'Verification checks Entra federation, foreign-subject rejection, and ACR token exchange, not repository access. Cloud IAM controls repository access; a component build proves access when pulling the image.',
}

export const ConnectionMetadata = ({
  connection,
}: {
  connection?: TCloudConnection
}) => {
  const values: [string, ReactNode][] = [
    [
      'Cloud',
      <CloudPlatform
        key="cloud"
        platform={connection?.platform || 'unknown'}
        displayVariant="name"
        variant="subtext"
      />,
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
      ? ([
          ['Tenant', connection.azure_tenant_id],
          ['Registry', connection.azure_registry_login_server],
          ['Registry resource ID', connection.azure_registry_id],
        ] as [string, string | undefined][])
      : ([
          ...(connection?.default_region
            ? [['Location', connection.default_region]]
            : []),
          ...(connection?.platform === 'gcp'
            ? [['Project number', connection.gcp_project_number]]
            : []),
        ] as [string, string | undefined][])),
  ]
  return (
    <>
      {values.map(([label, value]) => (
        <LabeledValue key={label} label={label} loading={!connection}>
          {typeof value === 'string' ? (
            <Text family="mono" variant="subtext" className="break-all">
              {value}
            </Text>
          ) : (
            value
          )}
        </LabeledValue>
      ))}
    </>
  )
}

export const ConnectionDetail = ({
  connection,
  orgId,
  basePath,
  tab,
  isVerifying = false,
  verificationTimedOut = false,
  onVerify,
  onDelete,
  error,
  installs,
  sharedConnections = [],
}: {
  connection?: TCloudConnection
  orgId: string
  basePath: string
  tab: TConnectionTab
  isVerifying?: boolean
  verificationTimedOut?: boolean
  onVerify: () => void
  onDelete: () => void
  error?: TAPIError | null
  installs?: ReactNode
  sharedConnections?: TCloudConnectionSummary[]
}) => (
  <DetailPage
    header={
      <DetailHeader
        backLink={false}
        loading={!connection && !error}
        title={connection?.name || 'Cloud connection'}
        description="Manage this connection and generate cloud setup commands."
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
        <PrincipalSharing connections={sharedConnections} orgId={orgId} />
        <Link href={`${basePath}/setup`}>View identity setup</Link>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <LabeledValue label="Created">
            <Time time={connection.created_at} format="long-datetime" />
          </LabeledValue>
          <LabeledValue label="Access">
            {connection.preset === 'stacks'
              ? 'Manages install stacks'
              : isImagePreset(connection.preset)
                ? 'Image repository access'
                : 'Custom policy'}
          </LabeledValue>
          {isImagePreset(connection.preset) && (
            <LabeledValue label="Image repositories">
              {connection.used_by?.image_repositories ?? 0}
            </LabeledValue>
          )}
        </div>
        <ConnectionPolicies connection={connection} orgId={orgId} />
        {isImagePreset(connection.preset) && (
          <Link href={`/${orgId}/settings/image-repositories`}>
            View image repositories
          </Link>
        )}
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
              ? isImagePreset(connection.preset)
                ? verifyingCopy[connection.preset]
                : 'Checking the OIDC exchange and account access.'
              : connection.status_message || 'Verification has not run yet.'}
          </LabeledValue>
        </div>
        {isImagePreset(connection.preset) && (
          <Text variant="subtext" theme="neutral">
            {imageNoteCopy[connection.preset]}
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
