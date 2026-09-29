import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { ClickToCopy } from '@/components/common/ClickToCopy'
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
import type { TAPIError, TCloudConnection } from '@/types'
import { ConnectionPolicies, failedRunbookStep } from '../ConnectionPolicies'
import { ConnectionStatus } from '../ConnectionStatus'

export type TConnectionTab = 'overview' | 'installs' | 'verification'

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
}) => (
  <DetailPage
    variant="page"
    header={
      <DetailHeader
        backLink={false}
        loading={!connection && !error}
        title={connection?.name || 'Cloud connection'}
        status={
          <ConnectionStatus connection={connection} isVerifying={isVerifying} />
        }
        identity={
          <ClickToCopy>
            <Text
              loading={!connection}
              family="mono"
              theme="neutral"
              variant="subtext"
              className="break-all"
            >
              {connection?.principal}
            </Text>
          </ClickToCopy>
        }
        metadata={
          <LabeledValue label="AWS account" loading={!connection}>
            <Text family="mono">
              {connection?.target_id}
              {connection?.default_region ? (
                <Text as="span" family="mono" theme="neutral">
                  {' '}
                  · {connection.default_region}
                </Text>
              ) : null}
            </Text>
          </LabeledValue>
        }
        actions={
          connection ? (
            <>
              <Button
                variant="primary"
                disabled={isVerifying}
                onClick={onVerify}
              >
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
        <Link href={`${basePath}/setup`}>View setup runbook</Link>
        <div className="grid grid-cols-2 gap-4">
          <LabeledValue label="Created">
            <Time time={connection.created_at} format="long-datetime" />
          </LabeledValue>
          <LabeledValue label="Access">
            {connection.preset === 'stacks'
              ? 'Manages install stacks'
              : 'Custom policy'}
          </LabeledValue>
        </div>
        <ConnectionPolicies connection={connection} />
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
              ? 'Checking the OIDC exchange and account access.'
              : connection.status_message || 'Verification has not run yet.'}
          </LabeledValue>
        </div>
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
