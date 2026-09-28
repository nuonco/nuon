import { useEffect, useRef, useState } from 'react'
import { DateTime } from 'luxon'
import { useSearchParams } from 'react-router'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Menu } from '@/components/common/Menu'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import {
  InstallsTableComponent,
  type InstallRow,
} from '@/components/installs/InstallsTable'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { PERMISSIONS_POLICY, TRUST_POLICY } from './mockData'
import type { TCloudConnectionOverview } from './overviewMockData'

export type TCloudConnectionDetailTab = 'overview' | 'installs' | 'verification'
export type TVerificationDisplay = 'current' | 'verifying'

const formatVerifiedTooltip = (time: string) => {
  const verifiedAt = DateTime.fromISO(time).toUTC()
  return `Verified ${verifiedAt.toRelative()} · ${verifiedAt.toFormat("LLL d, yyyy HH:mm 'UTC'")}`
}

export const ConnectionStatus = ({
  connection,
  status = connection.status,
}: {
  connection: TCloudConnectionOverview
  status?: TCloudConnectionOverview['status']
}) => {
  const tooltip =
    status === 'error'
      ? connection.statusMessage
      : status === 'verified' && connection.lastVerifiedAt
        ? formatVerifiedTooltip(connection.lastVerifiedAt)
        : connection.statusMessage || 'Verification is in progress.'

  return (
    <Tooltip
      position="bottom"
      tipContent={tooltip}
      tipContentClassName="max-w-xs whitespace-normal"
    >
      <span tabIndex={0}>
        <Status variant="badge" status={status} />
      </span>
    </Tooltip>
  )
}

const policyForConnection = (
  policy: string,
  connection: TCloudConnectionOverview
) =>
  policy
    .replaceAll('123456789012', connection.targetId)
    .replaceAll('us-east-1', connection.region)

const MOCK_INSTALL_ROWS: InstallRow[] = Array.from(
  { length: 14 },
  (_, index) => {
    const number = index + 1
    const updatedAt = DateTime.fromISO('2026-09-28T12:00:00Z')
      .minus({ hours: number })
      .toISO()!
    return {
      name: `acme-production-${number}`,
      nameHref: `/org-mock-001/installs/install-${number}`,
      installId: `install_01JEXAMPLE${String(number).padStart(2, '0')}`,
      appName: number % 2 ? 'Acme platform' : 'Acme data',
      appHref: `/org-mock-001/apps/app-${number % 2 ? 'platform' : 'data'}`,
      statuses: (
        <Status status={number === 6 ? 'pending' : 'active'} variant="badge" />
      ),
      region: (
        <Text family="mono" variant="subtext">
          {number % 3 ? 'us-east-1' : 'us-west-2'}
        </Text>
      ),
      platform: <Text variant="subtext">AWS</Text>,
      labels: (
        <Text family="mono" variant="label">
          env: production
        </Text>
      ),
      branch: (
        <Text family="mono" variant="subtext">
          main
        </Text>
      ),
      activity: <Time time={updatedAt} format="relative" variant="subtext" />,
      updatedAt,
      action: null,
    }
  }
)

const VERIFICATION_HISTORY = [
  {
    time: '2026-09-28T07:25:00Z',
    result: 'Verified',
    message: 'OIDC exchange and read-only CloudFormation probe passed.',
  },
  {
    time: '2026-09-27T16:12:00Z',
    result: 'Error',
    message: 'AWS denied cloudformation:DescribeStacks.',
  },
  {
    time: '2026-09-27T15:58:00Z',
    result: 'Verified',
    message: 'OIDC exchange and read-only CloudFormation probe passed.',
  },
  {
    time: '2026-09-26T09:40:00Z',
    result: 'Verified',
    message: 'OIDC exchange and read-only CloudFormation probe passed.',
  },
]

const HISTORY_COLUMNS = [
  {
    key: 'time' as const,
    header: 'Time',
    render: (value: unknown) => (
      <Time time={String(value)} format="relative" variant="subtext" />
    ),
  },
  {
    key: 'result' as const,
    header: 'Result',
    render: (value: unknown) => (
      <Status
        status={String(value).toLowerCase()}
        statusText={String(value)}
        variant="badge"
      />
    ),
  },
  {
    key: 'message' as const,
    header: 'Message',
    render: (value: unknown) => (
      <Text variant="subtext" theme="neutral">
        {String(value)}
      </Text>
    ),
  },
]

const OverviewTab = ({
  connection,
}: {
  connection: TCloudConnectionOverview
}) => (
  <div className="flex flex-col gap-6">
    <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
      <LabeledValue label="AWS account">
        <Text family="mono" variant="subtext">
          {connection.targetId}
        </Text>
      </LabeledValue>
      <LabeledValue label="Region">
        <Text family="mono" variant="subtext">
          {connection.region}
        </Text>
      </LabeledValue>
      <LabeledValue label="Created">
        <Time
          time={connection.createdAt}
          format="long-datetime"
          variant="subtext"
        />
      </LabeledValue>
      <LabeledValue label="Access">
        {connection.preset === 'Stacks'
          ? 'Manages install stacks'
          : 'Custom policy'}
      </LabeledValue>
    </div>

    <div className="flex flex-col gap-2">
      <Text variant="base" weight="strong">
        Trust policy
      </Text>
      <CodeBlock language="json" showCopy>
        {policyForConnection(TRUST_POLICY, connection)}
      </CodeBlock>
    </div>
    {connection.preset === 'Stacks' ? (
      <div className="flex flex-col gap-2">
        <Text variant="base" weight="strong">
          Permissions policy
        </Text>
        <CodeBlock language="json" showCopy>
          {policyForConnection(PERMISSIONS_POLICY, connection)}
        </CodeBlock>
      </div>
    ) : (
      <Banner theme="neutral">
        The permissions policy is managed in the customer's AWS account.
      </Banner>
    )}
  </div>
)

const InstallsTab = ({ empty = false }: { empty?: boolean }) => {
  const [searchParams] = useSearchParams()
  const limit = 10
  const offset = Number(searchParams.get('offset') || 0)
  const rows = empty ? [] : MOCK_INSTALL_ROWS
  return (
    <InstallsTableComponent
      data={rows.slice(offset, offset + limit)}
      isLoading={false}
      emptyTitle="No installs use this connection yet."
      emptyMessage="Select this connection when creating an install."
      pagination={{
        offset,
        limit,
        hasNext: offset + limit < rows.length,
      }}
    />
  )
}

const VerificationTab = ({
  connection,
  displayStatus,
  isVerifying,
  reverified,
  onReverify,
}: {
  connection: TCloudConnectionOverview
  displayStatus: TCloudConnectionOverview['status']
  isVerifying: boolean
  reverified: boolean
  onReverify: () => void
}) => {
  const verificationResult = isVerifying
    ? 'Checking the OIDC exchange and account access.'
    : reverified || connection.status === 'verified'
      ? connection.preset === 'Custom'
        ? 'OIDC exchange passed and connection ownership was confirmed.'
        : 'OIDC exchange and the read-only CloudFormation probe passed.'
      : connection.statusMessage || 'Verification has not run yet.'
  const failedAtTrust = connection.statusMessage?.includes('AssumeRole')

  return (
    <div className="flex flex-col gap-6">
      <div className="grid grid-cols-2 gap-4">
        <LabeledValue label="Last verified">
          {connection.lastVerifiedAt ? (
            <Time
              time={connection.lastVerifiedAt}
              format="long-datetime"
              variant="subtext"
            />
          ) : (
            <Text variant="subtext" theme="neutral">
              Never
            </Text>
          )}
        </LabeledValue>
        <LabeledValue label="Status">
          <ConnectionStatus connection={connection} status={displayStatus} />
        </LabeledValue>
        <LabeledValue label="Result" className="col-span-2">
          <Text variant="subtext" theme="neutral">
            {verificationResult}
          </Text>
        </LabeledValue>
      </div>
      {connection.status === 'error' && !isVerifying && !reverified && (
        <Banner theme="error">
          <div className="flex flex-col items-start gap-2">
            <Text variant="subtext">{connection.statusMessage}</Text>
            <Link href={failedAtTrust ? '#runbook-step-2' : '#runbook-step-3'}>
              View runbook step {failedAtTrust ? '2' : '3'}
            </Link>
          </div>
        </Banner>
      )}
      <Button
        className="w-fit"
        variant="secondary"
        disabled={isVerifying}
        onClick={onReverify}
      >
        {isVerifying && <Loading size={14} />}
        {isVerifying ? 'Re-verifying' : 'Re-verify'}
      </Button>
      <div className="flex flex-col gap-3">
        <Text variant="base" weight="strong">
          Verification history
        </Text>
        <PropertyGrid
          values={VERIFICATION_HISTORY}
          columns={HISTORY_COLUMNS}
          gridTemplate="max-content max-content minmax(0,1fr)"
          align="start"
        />
      </div>
    </div>
  )
}

interface ICloudConnectionDetailPage {
  connection: TCloudConnectionOverview
  selectedTab?: TCloudConnectionDetailTab
  installsEmpty?: boolean
  initialVerification?: TVerificationDisplay
}

export const CloudConnectionDetailPage = ({
  connection,
  selectedTab = 'overview',
  installsEmpty = false,
  initialVerification = 'current',
}: ICloudConnectionDetailPage) => {
  const [verification, setVerification] =
    useState<TVerificationDisplay>(initialVerification)
  const [reverified, setReverified] = useState(false)
  const timeoutRef = useRef<ReturnType<typeof setTimeout>>()
  const isVerifying = verification === 'verifying'
  const displayStatus = isVerifying
    ? 'pending'
    : reverified
      ? 'verified'
      : connection.status

  useEffect(
    () => () => {
      if (timeoutRef.current) clearTimeout(timeoutRef.current)
    },
    []
  )

  const reverify = () => {
    setVerification('verifying')
    setReverified(false)
    timeoutRef.current = setTimeout(() => {
      setVerification('current')
      setReverified(true)
    }, 1500)
  }

  const tabs = [
    { path: '/', text: 'Overview' },
    {
      path: '/installs',
      text: (
        <span className="flex items-center gap-2">
          Installs
          <Badge size="xs">
            {installsEmpty ? 0 : connection.installCount}
          </Badge>
        </span>
      ),
    },
    { path: '/verification', text: 'Verification' },
  ]
  const selectedIndex = ['overview', 'installs', 'verification'].indexOf(
    selectedTab
  )

  return (
    <DetailPage
      variant="page"
      header={
        <DetailHeader
          backLink={false}
          title={connection.name}
          status={
            <ConnectionStatus connection={connection} status={displayStatus} />
          }
          identity={
            <ClickToCopy className="max-w-full">
              <Text
                family="mono"
                variant="subtext"
                theme="neutral"
                className="min-w-0 max-w-2xl overflow-hidden text-ellipsis"
                nowrap
              >
                {connection.roleArn}
              </Text>
            </ClickToCopy>
          }
          actions={
            <>
              <Button
                variant="primary"
                disabled={isVerifying}
                onClick={reverify}
              >
                {isVerifying && <Loading size={14} />}
                {isVerifying ? 'Re-verifying' : 'Re-verify'}
              </Button>
              <Dropdown
                id="cloud-connection-actions"
                aria-label="Connection actions"
                buttonText={<Icon variant="DotsThreeIcon" size={20} />}
                hideIcon
                variant="ghost"
                buttonClassName="!p-2"
                alignment="right"
              >
                <Menu>
                  <Button variant="danger" onClick={() => {}}>
                    Delete connection
                  </Button>
                </Menu>
              </Dropdown>
            </>
          }
          metadata={
            <LabeledValue label="Cloud account">
              <Text family="mono" variant="subtext">
                {connection.targetId} · {connection.region}
              </Text>
            </LabeledValue>
          }
        />
      }
      tabNav={{
        basePath: '/cloud-connections/example',
        tabs,
        activeIndex: selectedIndex,
      }}
    >
      {selectedTab === 'overview' && <OverviewTab connection={connection} />}
      {selectedTab === 'installs' && <InstallsTab empty={installsEmpty} />}
      {selectedTab === 'verification' && (
        <VerificationTab
          connection={connection}
          displayStatus={displayStatus}
          isVerifying={isVerifying}
          reverified={reverified}
          onReverify={reverify}
        />
      )}
    </DetailPage>
  )
}
