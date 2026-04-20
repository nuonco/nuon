import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { IAMRolePoliciesCard, IAMRoleBoundaryExpand } from './IAMRoles'
import type { TInstallRole } from '@/lib/ctl-api/installs/get-latest-install-roles'

export const InstallRoleDetail = ({
  installRole,
}: {
  installRole: TInstallRole
}) => {
  const role = installRole.app_role_config
  if (!role) return null

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <Text weight="strong">Summary</Text>
        <div className="grid grid-cols-2 gap-6">
          <LabeledValue label="Created at">
            <Time
              variant="subtext"
              time={role.created_at}
              format="long-datetime"
            />
          </LabeledValue>
          <LabeledValue label="Name">{role.name}</LabeledValue>
          <LabeledValue label="Type">
            <Badge variant="code" size="sm">
              {role.type}
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
}
