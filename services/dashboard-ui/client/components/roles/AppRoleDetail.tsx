import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { IAMRolePoliciesCard, IAMRoleBoundaryExpand } from './IAMRoles'

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
  }[]
  permissions_boundary?: string
}

export const AppRoleDetail = ({ role }: { role: TAppRole }) => {
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
            <Badge variant="code" size="sm">
              {role?.type}
            </Badge>
          </LabeledValue>
        </div>
      </Card>

      <IAMRolePoliciesCard policies={role?.policies} />
      <IAMRoleBoundaryExpand permissionsBoundary={role?.permissions_boundary} />
    </div>
  )
}
