import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Modal } from '@/components/surfaces/Modal'
import type { TAppConfig } from '@/types'
import type {
  TInstallAppPermissionsConfig,
  TInstallPermissionsRoleStatus,
} from '@/lib/ctl-api/installs/get-install-app-permissions-config'
import { decodeAsString } from '@/utils/data-utils'

const IAMRoleBoundaryExpand = ({
  permissionsBoundary,
}: {
  permissionsBoundary?: string
}) => (
  <Expand
    id="permission-boundary"
    className="rounded-md border"
    heading={
      <Text weight="strong">
        Permission boundary{' '}
        <Text variant="subtext" weight="normal" theme="neutral">
          ({permissionsBoundary ? 'set' : 'not set'})
        </Text>
      </Text>
    }
    headerClassName="p-4"
  >
    <div className="p-4 border-t">
      {permissionsBoundary ? (
        <CodeBlock language="json">
          {decodeAsString(permissionsBoundary)}
        </CodeBlock>
      ) : (
        <Text>
          Set a permissions boundary to control the maximum permissions
          this role can have. This is not a common setting but can be
          used to delegate permission management to others.{' '}
          <Link
            className="!inline-flex"
            href="https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_boundaries.html?icmpid=docs_iam_console"
            isExternal
          >
            Learn more about permission boundaries{' '}
            <Icon variant="ArrowSquareOut" />
          </Link>
        </Text>
      )}
    </div>
  </Expand>
)

type TPolicy = {
  id?: string
  name?: string
  managed_policy_name?: string
  contents?: string
}

const IAMRolePoliciesCard = ({ policies }: { policies?: TPolicy[] }) => (
  <Card>
    <HeadingGroup>
      <Text weight="strong">
        Permissions policies{' '}
        <Text variant="subtext" weight="normal" theme="neutral">
          ({policies?.length})
        </Text>
      </Text>
      <Text variant="subtext" theme="neutral">
        You can attach up to 10 managed policies.
      </Text>
    </HeadingGroup>

    <div>
      <div className="grid grid-cols-3 gap-6 pb-2">
        <Text variant="subtext" theme="neutral">
          Policy name
        </Text>
        <Text variant="subtext" theme="neutral">
          Policy type
        </Text>
        <Text variant="subtext" theme="neutral">
          View more
        </Text>
      </div>
      {policies?.map((policy) => (
        <div
          key={policy?.id}
          className="grid grid-cols-3 gap-6 py-2 border-t"
        >
          {policy?.managed_policy_name ? (
            <>
              <Code variant="inline" className="!px-2">
                <Text variant="subtext" family="mono">
                  {policy?.managed_policy_name}
                </Text>
              </Code>

              <Text variant="subtext" weight="strong">
                AWS managed
              </Text>

              <Button
                className="!p-1"
                href={`https://docs.aws.amazon.com/aws-managed-policy/latest/reference/${policy?.managed_policy_name}.html`}
                size="sm"
              >
                <Icon variant="ArrowSquareOut" />
              </Button>
            </>
          ) : null}
          {policy?.contents ? (
            <>
              <Code variant="inline" className="!px-2">
                <Text variant="subtext" family="mono">
                  {policy?.name}
                </Text>
              </Code>

              <Text variant="subtext" weight="strong">
                Vendor defined
              </Text>

              <Modal
                size="half"
                heading={<>{policy?.name} policy JSON</>}
                triggerButton={{
                  className: '!p-1',
                  children: (
                    <span>
                      <Icon variant="BracketsCurly" />
                    </span>
                  ),
                  size: 'sm',
                }}
              >
                <div className="flex flex-col gap-2">
                  <ClickToCopyButton
                    className="!w-fit self-end"
                    textToCopy={decodeAsString(policy?.contents)}
                  />
                  <CodeBlock language="json">
                    {decodeAsString(policy?.contents)}
                  </CodeBlock>
                </div>
              </Modal>
            </>
          ) : null}
        </div>
      ))}
    </div>
  </Card>
)

export const IAMRoles = ({ appConfig }: { appConfig: TAppConfig }) => {
  return (
    <div className="flex flex-col divide-y gap-6">
      {appConfig?.permissions?.aws_iam_roles?.map((role) => (
        <div className="flex flex-col gap-4 pb-8" key={role?.id}>
          <div className="flex flex-col">
            <Text variant="h3" weight="strong">
              {role?.display_name}
            </Text>
            <Text variant="subtext" theme="neutral">
              {role?.description}
            </Text>
          </div>

          <Card>
            <Text weight="strong">Summary</Text>
            <div className="grid grid-cols-3 gap-6">
              <LabeledValue label="Created at">
                <Time
                  variant="subtext"
                  time={role?.created_at}
                  format="long-datetime"
                />
              </LabeledValue>
              <LabeledValue label="Name">{role?.name}</LabeledValue>
              <LabeledValue label="Type">
                <Badge variant="code" size="sm">
                  {role?.type}
                </Badge>
              </LabeledValue>
            </div>
          </Card>

          <IAMRolePoliciesCard policies={role?.policies} />
          <IAMRoleBoundaryExpand permissionsBoundary={role?.permissions_boundary} />
        </div>
      ))}
    </div>
  )
}

export const InstallIAMRoles = ({
  permissionsConfig,
}: {
  permissionsConfig: TInstallAppPermissionsConfig
}) => {
  const roles = [
    permissionsConfig.provision_role,
    permissionsConfig.deprovision_role,
    permissionsConfig.maintenance_role,
    ...(permissionsConfig.break_glass_roles ?? []),
    ...(permissionsConfig.custom_roles ?? []),
  ].filter(Boolean) as TInstallPermissionsRoleStatus[]

  return (
    <div className="flex flex-col divide-y gap-6">
      {roles.map((role) => (
        <div className="flex flex-col gap-4 pb-8" key={role.id}>
          <div className="flex flex-col">
            <Text variant="h3" weight="strong">
              {role.display_name}
            </Text>
            <Text variant="subtext" theme="neutral">
              {role.description}
            </Text>
          </div>

          <Card>
            <Text weight="strong">Summary</Text>
            <div className="grid grid-cols-5 gap-6">
              <LabeledValue label="Created at">
                <Time variant="subtext" time={role.created_at} format="long-datetime" />
              </LabeledValue>
              <LabeledValue label="Name">{role.name}</LabeledValue>
              <LabeledValue label="Type">
                <Badge variant="code" size="sm">
                  {role.type}
                </Badge>
              </LabeledValue>
              <LabeledValue label="Status">
                <Status status={role.enabled ? 'active' : 'inactive'}>
                  {role.enabled ? 'Provisioned' : 'Not provisioned'}
                </Status>
              </LabeledValue>
              <LabeledValue label="ARN">
                {role.enabled && role.arn ? (
                  <div className="flex items-center gap-1">
                    <Text variant="subtext" family="mono">
                      {role.arn}
                    </Text>
                    <ClickToCopyButton textToCopy={role.arn} />
                  </div>
                ) : (
                  <Text variant="subtext" theme="neutral">
                    —
                  </Text>
                )}
              </LabeledValue>
            </div>
          </Card>

          <IAMRolePoliciesCard policies={role.policies} />
          <IAMRoleBoundaryExpand permissionsBoundary={role.permissions_boundary} />
        </div>
      ))}
    </div>
  )
}

export const IAMRolesSkeleton = () => {
  return (
    <div className="flex flex-col divide-y">
      {Array.from({ length: 4 }).map((_, idx) => (
        <div className="flex flex-col gap-4 py-8" key={idx}>
          <div className="flex flex-col gap-1">
            <Skeleton width="250px" height="27px" />
            <Skeleton width="300px" height="17px" />
          </div>

          <Card>
            <Skeleton height="24px" width="65px" />
            <div className="grid grid-cols-3 gap-6">
              <LabeledValue label={<Skeleton height="17px" width="60px" />}>
                <Skeleton height="17px" width="205px" />
              </LabeledValue>
              <LabeledValue label={<Skeleton height="17px" width="32px" />}>
                <Skeleton height="17px" width="155px" />
              </LabeledValue>
              <LabeledValue label={<Skeleton height="17px" width="30px" />}>
                <Skeleton height="20px" width="130px" />
              </LabeledValue>
            </div>
          </Card>

          <Card>
            <Skeleton height="24px" width="150px" />
            <Skeleton height="17px" width="230px" />

            <div>
              <div className="grid grid-cols-3 gap-6 pb-2">
                <Skeleton height="17px" width="65px" />
                <Skeleton height="17px" width="60px" />
                <Skeleton height="17px" width="58px" />
              </div>

              <div className="grid grid-cols-3 gap-6 py-2 border-t">
                <Skeleton height="25px" width="150px" />
                <Skeleton height="17px" width="80px" />
                <Skeleton height="24px" width="24px" />
              </div>

              <div className="grid grid-cols-3 gap-6 py-2 border-t">
                <Skeleton height="25px" width="160px" />
                <Skeleton height="17px" width="85px" />
                <Skeleton height="24px" width="24px" />
              </div>
            </div>
          </Card>

          <Card>
            <Skeleton height="24px" width="190px" />
          </Card>
        </div>
      ))}
    </div>
  )
}
