import { Button } from '@/components/common/Button'
import { Select, type SelectOption } from '@/components/common/form/Select'
import { Text } from '@/components/common/Text'

const ROLE_TYPE_CONFIG = {
  custom: { theme: 'brand' as const, label: 'Custom' },
  break_glass: { theme: 'warn' as const, label: 'Break Glass' },
  provision: { theme: 'info' as const, label: 'Provision' },
  deprovision: { theme: 'neutral' as const, label: 'Deprovision' },
  maintenance: { theme: 'success' as const, label: 'Maintenance' },
}

type TAvailableRole = {
  name: string
  role_type: keyof typeof ROLE_TYPE_CONFIG
  default?: boolean
}

interface IRoleSelector {
  roles: TAvailableRole[]
  isLoading: boolean
  isError: boolean
  value?: string
  onChange?: (value: string) => void
  name?: string
  disabled?: boolean
}

export const RoleSelector = ({
  roles,
  isLoading,
  isError,
  value,
  onChange,
  name,
  disabled,
}: IRoleSelector) => {
  const defaultRole = roles.find((r) => r.default)

  const options: SelectOption[] = roles
    .filter((role) => !role.default)
    .map((role) => ({
      value: role.name,
      label: role.name,
      badge: {
        label: ROLE_TYPE_CONFIG[role.role_type]?.label,
        theme: ROLE_TYPE_CONFIG[role.role_type]?.theme,
      },
    }))

  const placeholder = defaultRole
    ? `${defaultRole.name}`
    : 'Use default role'

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
          <div className="flex items-center gap-2">
            <div className="flex-1">
              <Select
                name={name}
                value={value}
                onChange={(e) => onChange?.(e.target.value)}
                disabled={disabled}
                options={options}
                placeholder={placeholder}
              />
            </div>
            {value ? (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onChange?.('')}
              >
                Reset
              </Button>
            ) : null}
          </div>
        </>
      )}
    </div>
  )
}
