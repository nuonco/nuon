import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Modal } from '@/components/surfaces/Modal'
import type { TAppConfig } from '@/types'
import type {
  TInstallAppPermissionsConfig,
  TInstallPermissionsRoleStatus,
} from '@/lib/ctl-api/installs/get-install-app-permissions-config'
import type { TInstallRole } from '@/lib/ctl-api/installs/get-latest-install-roles'
import { decodeAsString } from '@/utils/data-utils'
import { humanize } from '@/utils/string-utils'

export const IAMRoleBoundaryExpand = ({
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
            Learn more about permission boundaries
          </Link>
        </Text>
      )}
    </div>
  </Expand>
)

export type TPolicy = {
  id?: string
  name?: string
  managed_policy_name?: string
  contents?: string
  gcp_predefined_role?: string
  gcp_permissions?: string[]
  azure_built_in_roles?: string[]
  azure_actions?: string[]
}

export const IAMRolePoliciesCard = ({ policies }: { policies?: TPolicy[] }) => (
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

              <Link
                href={`https://docs.aws.amazon.com/aws-managed-policy/latest/reference/${policy?.managed_policy_name}.html`}
                isExternal
              >
                View docs
              </Link>
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
                size="sm"
                heading={<>{policy?.name} policy JSON</>}
                triggerButton={{
                  className: '!p-1',
                  children: (
                    <span>
                      <Icon variant="BracketsCurlyIcon" />
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
          {policy?.gcp_predefined_role ? (
            <>
              <Code variant="inline" className="!px-2">
                <Text variant="subtext" family="mono">
                  {policy?.gcp_predefined_role}
                </Text>
              </Code>

              <Text variant="subtext" weight="strong">
                GCP predefined
              </Text>

              <Link
                href="https://cloud.google.com/iam/docs/understanding-roles"
                isExternal
              >
                View docs
              </Link>
            </>
          ) : null}
          {policy?.gcp_permissions?.length ? (
            <>
              <Code variant="inline" className="!px-2">
                <Text variant="subtext" family="mono">
                  {policy?.name}
                </Text>
              </Code>

              <Text variant="subtext" weight="strong">
                GCP custom
              </Text>

              <Modal
                size="sm"
                heading={<>{policy?.name} permissions</>}
                triggerButton={{
                  className: '!p-1',
                  children: (
                    <span>
                      <Icon variant="BracketsCurlyIcon" />
                    </span>
                  ),
                  size: 'sm',
                }}
              >
                <div className="flex flex-col gap-2">
                  <ClickToCopyButton
                    className="!w-fit self-end"
                    textToCopy={policy.gcp_permissions.join('\n')}
                  />
                  <CodeBlock language="text">
                    {policy.gcp_permissions.join('\n')}
                  </CodeBlock>
                </div>
              </Modal>
            </>
          ) : null}
          {policy?.azure_built_in_roles?.length ? (
            <>
              <Code variant="inline" className="!px-2">
                <Text variant="subtext" family="mono">
                  {policy.azure_built_in_roles.join(', ')}
                </Text>
              </Code>

              <Text variant="subtext" weight="strong">
                Azure built-in
              </Text>

              <Link
                href="https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles"
                isExternal
              >
                View docs
              </Link>
            </>
          ) : null}
          {policy?.azure_actions?.length ? (
            <>
              <Code variant="inline" className="!px-2">
                <Text variant="subtext" family="mono">
                  {policy?.name}
                </Text>
              </Code>

              <Text variant="subtext" weight="strong">
                Azure custom
              </Text>

              <Modal
                size="sm"
                heading={<>{policy?.name} actions</>}
                triggerButton={{
                  className: '!p-1',
                  children: (
                    <span>
                      <Icon variant="BracketsCurlyIcon" />
                    </span>
                  ),
                  size: 'sm',
                }}
              >
                <div className="flex flex-col gap-2">
                  <ClickToCopyButton
                    className="!w-fit self-end"
                    textToCopy={policy.azure_actions.join('\n')}
                  />
                  <CodeBlock language="text">
                    {policy.azure_actions.join('\n')}
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
            <Text variant="h3" weight="strong" id={role?.display_name}>
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
                <Badge size="sm">
                  {humanize(role?.type)}
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
  installRoles,
}: {
  installRoles: TInstallRole[]
}) => {
  return (
    <div className="flex flex-col divide-y gap-6">
      {installRoles.map((installRole) => {
        const role = installRole.app_role_config
        if (!role) return null

        return (
          <div className="flex flex-col gap-4 pb-8" key={installRole.id}>
            <div className="flex flex-col">
              <Text variant="h3" weight="strong" level={3} role="heading" id={role?.display_name}>
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
                  <Badge size="sm">
                    {humanize(role.type)}
                  </Badge>
                </LabeledValue>
                <LabeledValue label="Status">
                  <Status status={installRole.provisioned ? 'active' : 'inactive'}>
                    {installRole.provisioned ? 'Provisioned' : 'Not provisioned'}
                  </Status>
                </LabeledValue>
                <LabeledValue label="ARN">
                  {installRole.role_id ? (
                    <div className="flex items-start gap-1 min-w-0">
                      <Text variant="subtext" family="mono" className="break-all">
                        {installRole.role_id}
                      </Text>
                      <ClickToCopyButton textToCopy={installRole.role_id} />
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
        )
      })}
    </div>
  )
}
