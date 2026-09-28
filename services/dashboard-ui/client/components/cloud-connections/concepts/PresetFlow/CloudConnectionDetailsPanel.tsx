import { useEffect, useRef, useState } from 'react'
import { DateTime } from 'luxon'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Divider } from '@/components/common/Divider'
import { EmptyState } from '@/components/common/EmptyState'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { PERMISSIONS_POLICY, TRUST_POLICY } from './mockData'
import type {
  TCloudConnectionInstall,
  TCloudConnectionOverview,
} from './overviewMockData'

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

const INSTALL_COLUMNS = [
  {
    key: 'name' as const,
    header: 'Install',
    render: (_: unknown, install: TCloudConnectionInstall) => (
      <Text variant="subtext" weight="strong">
        {install.name}
      </Text>
    ),
  },
  {
    key: 'appName' as const,
    header: 'App',
    render: (_: unknown, install: TCloudConnectionInstall) => (
      <Text variant="subtext">{install.appName}</Text>
    ),
  },
  {
    key: 'status' as const,
    header: 'Status',
    render: (_: unknown, install: TCloudConnectionInstall) => (
      <Status status={install.status} variant="badge" />
    ),
  },
]

interface ICloudConnectionDetailsPanel extends IPanel {
  connection: TCloudConnectionOverview
  initialSection?: 'summary' | 'policy' | 'verification'
  initialVerification?: TVerificationDisplay
}

export const CloudConnectionDetailsPanel = ({
  connection,
  initialSection = 'summary',
  initialVerification = 'current',
  ...panelProps
}: ICloudConnectionDetailsPanel) => {
  const [verification, setVerification] =
    useState<TVerificationDisplay>(initialVerification)
  const [reverified, setReverified] = useState(false)
  const policyRef = useRef<HTMLDivElement>(null)
  const verificationRef = useRef<HTMLDivElement>(null)
  const timeoutRef = useRef<ReturnType<typeof setTimeout>>()
  const isVerifying = verification === 'verifying'
  const displayStatus = isVerifying
    ? 'pending'
    : reverified
      ? 'verified'
      : connection.status

  useEffect(() => {
    if (initialSection === 'summary') return
    const timeout = setTimeout(
      () =>
        (initialSection === 'policy'
          ? policyRef.current
          : verificationRef.current
        )?.scrollIntoView({ block: 'start' }),
      200
    )
    return () => clearTimeout(timeout)
  }, [initialSection])

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

  const verificationResult = isVerifying
    ? 'Checking the OIDC exchange and account access.'
    : reverified || connection.status === 'verified'
      ? connection.preset === 'Custom'
        ? 'OIDC exchange passed and connection ownership was confirmed.'
        : 'OIDC exchange and the read-only CloudFormation probe passed.'
      : connection.statusMessage || 'Verification has not run yet.'

  const failedAtTrust = connection.statusMessage?.includes('AssumeRole')

  return (
    <Panel heading={connection.name} size="half" {...panelProps}>
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Summary
        </Text>
        <div className="grid grid-cols-2 gap-4">
          <LabeledValue label="Name">{connection.name}</LabeledValue>
          <LabeledValue label="Status">
            <div className="flex flex-wrap items-center gap-2">
              <ConnectionStatus
                connection={connection}
                status={displayStatus}
              />
              {connection.lastVerifiedAt && !isVerifying && (
                <Time
                  time={connection.lastVerifiedAt}
                  format="relative"
                  variant="subtext"
                  theme="neutral"
                />
              )}
            </div>
          </LabeledValue>
          <LabeledValue label="Role ARN" className="col-span-2 min-w-0">
            <ClickToCopy className="max-w-full">
              <Text family="mono" variant="subtext" className="block truncate">
                {connection.roleArn}
              </Text>
            </ClickToCopy>
          </LabeledValue>
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
      </div>

      <div ref={policyRef} className="scroll-mt-6">
        <Divider dividerWord="Policy" />
      </div>
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-2">
          <Text weight="strong">Trust policy</Text>
          <CodeBlock language="json" showCopy>
            {policyForConnection(TRUST_POLICY, connection)}
          </CodeBlock>
        </div>
        {connection.preset === 'Stacks' ? (
          <div className="flex flex-col gap-2">
            <Text weight="strong">Permissions policy</Text>
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

      <Divider dividerWord="Installs using this connection" />
      {connection.installs.length ? (
        <PropertyGrid
          values={connection.installs}
          columns={INSTALL_COLUMNS}
          gridTemplate="minmax(0,1.2fr) minmax(0,1fr) max-content"
        />
      ) : (
        <EmptyState
          size="sm"
          variant="table"
          emptyTitle="No installs use this connection yet."
          emptyMessage="Select this connection when creating an install."
        />
      )}

      <div ref={verificationRef} className="scroll-mt-6">
        <Divider dividerWord="Verification" />
      </div>
      <div className="flex flex-col gap-4">
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
              <Link
                href={failedAtTrust ? '#runbook-step-2' : '#runbook-step-3'}
              >
                View runbook step {failedAtTrust ? '2' : '3'}
              </Link>
            </div>
          </Banner>
        )}
        <Button
          className="w-fit"
          variant="secondary"
          disabled={isVerifying}
          onClick={reverify}
        >
          {isVerifying && <Loading size={14} />}
          {isVerifying ? 'Re-verifying' : 'Re-verify'}
        </Button>
      </div>
    </Panel>
  )
}
