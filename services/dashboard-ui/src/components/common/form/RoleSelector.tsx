'use client'

import { useEffect, useState } from 'react'
import { Select, type SelectOption } from '@/components/common/form/Select'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import type { AvailableRole } from '@/lib'

export type OperationType =
  | 'provision'
  | 'reprovision'
  | 'deprovision'
  | 'deploy'
  | 'teardown'
  | 'trigger'

export type PrincipalType = 'component' | 'sandbox' | 'action'

interface IRoleSelector {
  installId: string
  operationType: OperationType
  principalType: PrincipalType
  value?: string
  onChange?: (e: React.ChangeEvent<HTMLSelectElement>) => void
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
  const [roles, setRoles] = useState<AvailableRole[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function fetchRoles() {
      if (!installId || !org?.id) return

      setIsLoading(true)
      setError(null)

      try {
        const url = `/api/orgs/${org.id}/installs/${installId}/available-roles?operation_type=${operationType}&principal_type=${principalType}`

        const response = await fetch(url)

        if (!response.ok) {
          const errorData = await response.json()
          setError('Failed to load available roles')
          setIsLoading(false)
          return
        }

        const result = await response.json()

        if (result.data?.roles) {
          setRoles(result.data.roles)
        } 
      } catch (err) {
        setError('Failed to load available roles')
      }

      setIsLoading(false)
    }

    fetchRoles()
  }, [installId, operationType, principalType, org?.id])

  // Convert roles to select options
  const options: SelectOption[] = roles.map((role) => ({
    value: role.name,
    label: role.name,
    badge: {
      label: ROLE_TYPE_CONFIG[role.role_type]?.label,
      theme: ROLE_TYPE_CONFIG[role.role_type]?.theme,
    },
  }))

  if (isLoading) {
    return (
      <div className="flex flex-col gap-1">
        <Text variant="label" weight="strong">
          Execution Role (optional)
        </Text>
        <Text variant="subtext" theme="neutral">
          Loading available roles...
        </Text>
      </div>
    )
  }

  if (error || roles.length === 0) {
    return null
  }

  return (
    <div className="flex flex-col gap-1">
      <Text variant="label" weight="strong">
        Execution Role (optional)
      </Text>
      <Text variant="subtext" theme="neutral">
        Select an IAM role to use for this operation. If not selected, the
        default role will be used.
      </Text>
      <Select
        name={name}
        value={value || ''}
        onChange={onChange}
        disabled={disabled}
        options={options}
        placeholder="Use default role"
      />
    </div>
  )
}
