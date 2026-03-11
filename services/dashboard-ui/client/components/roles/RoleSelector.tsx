import { useQuery } from '@tanstack/react-query'
import { Select, type SelectOption } from '@/components/common/form/Select'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { getAvailableRoles } from '@/lib'
import type { TOperationType, TPrincipalType } from '@/types'

interface IRoleSelector {
  installId: string
  operationType: TOperationType
  principalType: TPrincipalType
  value?: string
  onChange?: (value: string) => void
  name?: string
  disabled?: boolean
}

const ROLE_TYPE_CONFIG = {
  custom: { theme: 'brand' as const, label: 'Custom' },
  break_glass: { theme: 'warn' as const, label: 'Break Glass' },
  provision: { theme: 'info' as const, label: 'Provision' },
  deprovision: { theme: 'neutral' as const, label: 'Deprovision' },
  maintenance: { theme: 'success' as const, label: 'Maintenance' },
}

export const RoleSelector = ({
  installId,
  operationType,
  principalType,
  value,
  onChange,
  name,
  disabled,
}: IRoleSelector) => {
  const { org } = useOrg()

  const { data, isLoading, isError } = useQuery({
    queryKey: ['available-roles', org.id, installId, operationType, principalType],
    queryFn: () =>
      getAvailableRoles({ installId, operationType, principalType, orgId: org.id }),
    enabled: !!installId && !!org.id,
  })

  const roles = data?.roles ?? []

  const options: SelectOption[] = roles.map((role) => ({
    value: role.name,
    label: role.name,
    badge: {
      label: ROLE_TYPE_CONFIG[role.role_type]?.label,
      theme: ROLE_TYPE_CONFIG[role.role_type]?.theme,
    },
  }))

  return (
    <div className="flex flex-col gap-1">
      <Text variant="label" weight="strong">
        Execution Role (optional)
      </Text>
      {isLoading ? (
        <Text variant="subtext" theme="neutral">
          Loading available roles...
        </Text>
      ) : isError ? (
        <Text variant="subtext" theme="neutral">
          Failed to load available roles
        </Text>
      ) : roles.length === 0 ? (
        <Text variant="subtext" theme="neutral">
          No roles available from install stack outputs
        </Text>
      ) : (
        <>
          <Text variant="subtext" theme="neutral">
            Select an IAM role to use for this operation. If not selected, the
            default role will be used.
          </Text>
          <Select
            name={name}
            value={value}
            onChange={(e) => onChange?.(e.target.value)}
            disabled={disabled}
            options={options}
            placeholder="Use default role"
          />
        </>
      )}
    </div>
  )
}
