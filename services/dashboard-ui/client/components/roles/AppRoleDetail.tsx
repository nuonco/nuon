import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import {
  IAMRoleBoundaryExpand,
  IAMRoleNamedPoliciesExpand,
  IAMRolePoliciesCard,
} from './IAMRoles'
import { humanize } from '@/utils/string-utils'
import type { TNamedIAMPolicy } from '@/lib/ctl-api/installs/get-install-app-permissions-config'

type TAppRole = {
  id?: string
  display_name?: string
  description?: string
  name?: string
  type?: string
  created_at?: string
  policies?: {
    id?: string
    name?: string
    managed_policy_name?: string
    contents?: string
    gcp_predefined_role?: string
    gcp_permissions?: string[]
    azure_built_in_roles?: string[]
    azure_actions?: string[]
  }[]
  named_policy_names?: string[]
  permissions_boundary?: string
}

export const AppRoleDetail = ({
  role,
  namedPolicies = [],
}: {
  role: TAppRole
  namedPolicies?: TNamedIAMPolicy[]
}) => {
  const attachedNamedPolicies = namedPolicies.filter((policy) =>
    role.named_policy_names?.includes(policy.name ?? '')
  )

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <Text weight="strong">Summary</Text>
        <div className="grid grid-cols-2 gap-6">
          <LabeledValue label="Created at">
            <Time
              variant="subtext"
              time={role?.created_at}
              format="long-datetime"
            />
          </LabeledValue>
          <LabeledValue label="Name">{role?.name}</LabeledValue>
          <LabeledValue label="Type">
            <Badge size="sm">{humanize(role?.type)}</Badge>
          </LabeledValue>
        </div>
      </Card>

      <IAMRolePoliciesCard policies={role?.policies} />
      <IAMRoleNamedPoliciesExpand
        id={role.id ?? role.name ?? 'app-role'}
        policies={attachedNamedPolicies}
      />
      <IAMRoleBoundaryExpand permissionsBoundary={role?.permissions_boundary} />
    </div>
  )
}
