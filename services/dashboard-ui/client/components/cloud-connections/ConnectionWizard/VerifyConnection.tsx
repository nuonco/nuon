import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TAPIError, TCloudConnection } from '@/types'
import { identityConsoleURL, failedRunbookStep } from '../ConnectionPolicies'
import { isImagePreset } from '../repository-scope'

const platformCopy = {
  azure: {
    description:
      'Nuon checks Entra federation and foreign-subject rejection, and exchanges a token with ACR. Component builds check repository access.',
    checking:
      'Checking the connection. Entra federation and ACR role changes can take a few minutes to propagate.',
    verified:
      'Connection verified. Entra federation, foreign-subject rejection, and ACR token exchange passed.',
    console: 'View application in Azure portal',
    failure: (step: number) =>
      step === 3
        ? 'ACR token exchange failed'
        : 'Entra federation trust check failed',
    checks: [
      'Exchange this connection’s OIDC token through Microsoft Entra federation',
      'Reject a foreign connection subject',
      'Exchange the Entra token with ACR',
    ],
  },
  gcp: {
    description:
      'Nuon exchanges an OIDC token through GCP STS, impersonates the service account, and rejects a foreign subject. Component builds check repository access.',
    checking:
      'Checking the connection. IAM changes can take up to a minute to propagate.',
    verified:
      'Connection verified. Federation, service account impersonation, and foreign-subject rejection passed.',
    console: 'View service accounts in GCP console',
    failure: () => 'Federation trust check failed',
    checks: [
      'Exchange this connection’s OIDC token through GCP STS',
      'Impersonate the selected service account',
      'Reject a foreign connection subject',
    ],
  },
}

const awsCopy: Record<
  string,
  {
    description: string
    verified: string
    failure: string
    extraChecks: string[]
  }
> = {
  custom: {
    description:
      'Nuon checks that the role trusts this connection and rejects other Nuon connections.',
    verified:
      'Connection verified. Role trust and foreign-subject rejection passed; custom permissions were not checked.',
    failure: 'CloudFormation probe failed',
    extraChecks: [],
  },
  ecr: {
    description:
      'Nuon checks the role trust, rejects other Nuon connections, and retrieves an ECR token. Component builds check repository access.',
    verified:
      'Connection verified. Role trust, foreign-subject rejection, and ECR token retrieval passed.',
    failure: 'ECR token retrieval failed',
    extraChecks: ['Retrieve an ECR token with ecr:GetAuthorizationToken'],
  },
  stacks: {
    description:
      'Nuon checks the role trust, rejects other Nuon connections, and runs a read-only CloudFormation probe.',
    verified:
      'Connection verified. Role trust, foreign-subject rejection, and the read-only CloudFormation probe passed.',
    failure: 'CloudFormation probe failed',
    extraChecks: ['Read the install stack with cloudformation:DescribeStacks'],
  },
}

const awsChecks = [
  'Exchange an OIDC token for this connection',
  'Reject another Nuon connection',
]

export const VerifyConnection = ({
  connection,
  isVerifying,
  verificationTimedOut = false,
  error,
  onVerify,
  setupHref,
  detailHref,
}: {
  connection: TCloudConnection
  isVerifying: boolean
  verificationTimedOut?: boolean
  error?: TAPIError | null
  onVerify: () => void
  setupHref: string
  detailHref: string
}) => {
  const verified =
    !isVerifying &&
    !verificationTimedOut &&
    !error &&
    connection.status === 'verified'
  const failed =
    !isVerifying &&
    !verificationTimedOut &&
    (error || connection.status === 'error')
  const message =
    error?.description || error?.error || connection.status_message
  const step = failedRunbookStep({
    ...connection,
    status_message: error
      ? [error.error, error.description].filter(Boolean).join(' ')
      : message,
  })
  const copy = platformCopy[connection.platform]
  const aws = awsCopy[connection.preset] ?? awsCopy.stacks
  return (
    <div className="flex max-w-3xl flex-col gap-6">
      <SectionHeader
        title="Verify the connection"
        description={copy ? copy.description : aws.description}
      />
      {isVerifying && (
        <Banner theme="info">
          {copy
            ? copy.checking
            : 'Checking the connection. IAM changes can take up to a minute to propagate.'}
        </Banner>
      )}
      {verificationTimedOut && (
        <Banner theme="info">Still checking — refresh in a moment</Banner>
      )}
      {verified && (
        <Banner theme="success">{copy ? copy.verified : aws.verified}</Banner>
      )}
      {failed && (
        <Banner theme="error">
          <div className="flex flex-col items-start gap-2">
            <Text weight="strong">
              {copy
                ? copy.failure(step)
                : error || message?.startsWith('Verification failed:')
                  ? 'Verification failed'
                  : step === 1
                    ? 'OIDC identity check failed'
                    : step === 2
                      ? 'Role trust check failed'
                      : aws.failure}
            </Text>
            <Text>{message}</Text>
            <Link href={`${setupHref}#runbook-step-${step}`}>
              View runbook step {step}
            </Link>
            <Link href={identityConsoleURL(connection)} isExternal>
              {copy ? copy.console : 'View role in AWS console'}
            </Link>
          </div>
        </Banner>
      )}
      <div className="flex flex-col gap-3" aria-label="Verification checks">
        <Text weight="strong">Verification checks</Text>
        {(copy ? copy.checks : [...awsChecks, ...aws.extraChecks]).map(
          (label) => (
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
          )
        )}
      </div>
      {isImagePreset(connection.preset) && (
        <Text variant="subtext" theme="neutral">
          Verification does not check repository permissions. Cloud IAM controls
          access; a component build proves image access when it pulls.
        </Text>
      )}
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
