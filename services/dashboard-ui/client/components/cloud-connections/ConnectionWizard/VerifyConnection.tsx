import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TAPIError, TCloudConnection } from '@/types'
import { awsRoleURL, failedRunbookStep } from '../ConnectionPolicies'

export const VerifyConnection = ({
  connection,
  isVerifying,
  error,
  onVerify,
  setupHref,
  detailHref,
}: {
  connection: TCloudConnection
  isVerifying: boolean
  error?: TAPIError | null
  onVerify: () => void
  setupHref: string
  detailHref: string
}) => {
  const verified = !isVerifying && !error && connection.status === 'verified'
  const failed = !isVerifying && (error || connection.status === 'error')
  const message =
    error?.description || error?.error || connection.status_message
  const step = failedRunbookStep({ ...connection, status_message: message })
  return (
    <div className="flex max-w-3xl flex-col gap-6">
      <SectionHeader
        title="Verify the connection"
        description={
          connection.preset === 'custom'
            ? 'Nuon checks that the role trusts this connection and rejects other Nuon connections.'
            : 'Nuon checks the role trust, rejects other Nuon connections, and runs a read-only CloudFormation probe.'
        }
      />
      {isVerifying && (
        <Banner theme="info">
          Verifying connection. Testing the OIDC exchange and account access.
        </Banner>
      )}
      {verified && (
        <Banner theme="success">
          Connection verified. The role is limited to this connection and the
          selected access is available.
        </Banner>
      )}
      {failed && (
        <Banner theme="error">
          <div className="flex flex-col items-start gap-2">
            <Text weight="strong">
              {error
                ? 'Verification failed'
                : step === 2
                  ? 'Role trust check failed'
                  : 'CloudFormation probe failed'}
            </Text>
            <Text>{message}</Text>
            <Link href={`${setupHref}#runbook-step-${step}`}>
              View runbook step {step}
            </Link>
            <Link href={awsRoleURL(connection.principal)} isExternal>
              Open in AWS console
            </Link>
          </div>
        </Banner>
      )}
      <Card>
        <Text weight="strong">Verification checks</Text>
        {[
          'Exchange an OIDC token for this connection',
          'Reject another Nuon connection',
          ...(connection.preset === 'stacks'
            ? ['Read the install stack with cloudformation:DescribeStacks']
            : []),
        ].map((label) => (
          <div key={label} className="flex items-center gap-3">
            {isVerifying ? (
              <Loading size={18} />
            ) : (
              <Icon
                variant={verified ? 'CheckCircleIcon' : 'ClockIcon'}
                theme={verified ? 'success' : 'neutral'}
              />
            )}
            <Text>{label}</Text>
          </div>
        ))}
      </Card>
      <div className="flex flex-wrap gap-3">
        {!verified && (
          <Button variant="primary" disabled={isVerifying} onClick={onVerify}>
            {isVerifying ? 'Verifying connection' : 'Verify connection'}
          </Button>
        )}
        <Button variant={verified ? 'primary' : 'secondary'} href={detailHref}>
          View connection
        </Button>
      </div>
    </div>
  )
}
