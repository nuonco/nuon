export default {
  title: 'Roles/RoleSelector',
}

import { useState } from 'react'
import { RoleSelector } from './RoleSelector'

const mockRoles = [
  { name: 'default-provision', role_type: 'provision' as const, default: true },
  { name: 'custom-deploy', role_type: 'custom' as const },
  { name: 'break-glass-admin', role_type: 'break_glass' as const },
  { name: 'maintenance-ops', role_type: 'maintenance' as const },
]

export const Default = () => {
  const [value, setValue] = useState('')
  return (
    <RoleSelector
      roles={mockRoles}
      isLoading={false}
      isError={false}
      value={value}
      onChange={setValue}
    />
  )
}

export const Loading = () => (
  <RoleSelector roles={[]} isLoading={true} isError={false} />
)

export const Error = () => (
  <RoleSelector roles={[]} isLoading={false} isError={true} />
)

export const Empty = () => (
  <RoleSelector roles={[]} isLoading={false} isError={false} />
)

export const WithSelectedValue = () => (
  <RoleSelector
    roles={mockRoles}
    isLoading={false}
    isError={false}
    value="custom-deploy"
    onChange={() => {}}
  />
)

export const WorkflowDefaults = () => (
  <div className="flex flex-col gap-8 max-w-xl">
    {(['maintenance', 'provision', 'deprovision'] as const).map((roleType) => (
      <RoleSelector
        key={roleType}
        roles={[
          ...mockRoles.map((role) => ({
            ...role,
            default: role.role_type === roleType,
          })),
          {
            name: 'default-deprovision',
            role_type: 'deprovision',
            default: roleType === 'deprovision',
          },
        ]}
        isLoading={false}
        isError={false}
        isWorkflowDefault
      />
    ))}
  </div>
)
