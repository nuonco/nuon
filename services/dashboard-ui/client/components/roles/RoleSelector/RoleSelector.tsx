import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { Select, type SelectOption } from '@/components/common/form/Select'

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
  hideWhenEmpty?: boolean
}

export const RoleSelector = ({
  roles,
  isLoading,
  isError,
  value,
  onChange,
  name,
  disabled,
  hideWhenEmpty,
}: IRoleSelector) => {
  if (!isLoading && !isError && roles.length === 0) {
    if (hideWhenEmpty) {
      return null
    }
    return (
      <Banner theme="warn">
        <Text variant="body">
          No execution roles are available for this operation. Add or enable a role in your install stack to continue.
        </Text>
      </Banner>
    )
  }

  const defaultRole = roles.find((r) => r.default)

  const roleOption = (role: TAvailableRole, value: string): SelectOption => ({
    value,
    label: role.name,
    badge: {
      label: ROLE_TYPE_CONFIG[role.role_type]?.label,
      theme: ROLE_TYPE_CONFIG[role.role_type]?.theme,
    },
  })

  const options: SelectOption[] = [
    ...(defaultRole ? [roleOption(defaultRole, '')] : []),
    ...roles.filter((role) => !role.default).map((role) => roleOption(role, role.name)),
  ]

  const helperText = isLoading
    ? 'Loading available roles…'
    : isError
      ? 'Failed to load available roles'
      : 'If unset, the default role is used.'

  return (
    <Select
      name={name}
      value={value ?? ''}
      onChange={(e) => onChange?.(e.target.value)}
      disabled={disabled || isLoading || isError}
      options={options}
      placeholder={defaultRole ? defaultRole.name : 'Use default role'}
      labelProps={{ labelText: 'Execution role (optional)' }}
      helperText={helperText}
      helperTextProps={{
        variant: 'subtext',
        className: 'text-cool-grey-500 dark:text-cool-grey-400',
      }}
    />
  )
}
