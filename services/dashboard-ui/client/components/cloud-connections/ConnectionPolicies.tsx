import { Banner } from '@/components/common/Banner'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import type { TCloudConnection, TCloudConnectionSetup } from '@/types'

export const roleName = (principal: string) =>
  principal.split(':role/')[1] || ''
export const awsRoleURL = (principal: string) =>
  `https://console.aws.amazon.com/iam/home#/roles/details/${encodeURIComponent(roleName(principal))}?section=trust`
export const failedRunbookStep = (connection: TCloudConnection) =>
  connection.preset === 'stacks' &&
  /CloudFormation/i.test(connection.status_message || '')
    ? 3
    : 2

export const ConnectionPolicies = ({
  connection,
  setup,
}: {
  connection: TCloudConnection
  setup?: TCloudConnectionSetup
}) => {
  if (!setup) return <Loading />
  return (
    <div className="flex min-w-0 flex-col gap-6">
      <section
        id="runbook-step-2"
        className="flex min-w-0 flex-col gap-3 scroll-mt-6"
      >
        <div className="flex items-center justify-between gap-3">
          <Text weight="strong">Trust policy</Text>
          <Link href={awsRoleURL(connection.principal)} isExternal>
            Open in AWS console
          </Link>
        </div>
        <Text variant="subtext" theme="neutral">
          Copy this policy as-is. To customize the role beyond it, edit it in
          the IAM console after creating it — the audience and subject
          conditions must stay unchanged for verification to pass.
        </Text>
        <CodeBlock language="json" showCopy>
          {JSON.stringify(setup.trust_policy, null, 2)}
        </CodeBlock>
      </section>
      {connection.preset === 'stacks' ? (
        <section
          id="runbook-step-3"
          className="flex min-w-0 flex-col gap-3 scroll-mt-6"
        >
          <Text weight="strong">Permissions policy</Text>
          <CodeBlock language="json" showCopy>
            {JSON.stringify(setup.permissions_policy, null, 2)}
          </CodeBlock>
        </section>
      ) : (
        <Banner theme="neutral">
          The permissions policy is managed in the customer's AWS account.
        </Banner>
      )}
    </div>
  )
}
