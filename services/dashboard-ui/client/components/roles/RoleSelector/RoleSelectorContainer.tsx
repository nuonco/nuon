import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getAvailableRoles } from '@/lib'
import type { TOperationType, TPrincipalType } from '@/types'
import { RoleSelector } from './RoleSelector'

interface IRoleSelectorContainer {
  installId: string
  operationType?: TOperationType
  principalType?: TPrincipalType
  principalId?: string
  value?: string
  onChange?: (value: string) => void
  name?: string
  disabled?: boolean
}

export const RoleSelectorContainer = ({
  installId,
  operationType,
  principalType,
  principalId,
  value,
  onChange,
  name,
  disabled,
}: IRoleSelectorContainer) => {
  const { org } = useOrg()

  const { data, isLoading, isError } = useQuery({
    queryKey: ['available-roles', org.id, installId, operationType, principalType, principalId],
    queryFn: () =>
      getAvailableRoles({ installId, operationType, principalType, principalId, orgId: org.id }),
    enabled: !!installId && !!org.id,
  })

  const roles = data?.roles ?? []

  return (
    <RoleSelector
      roles={roles as any}
      isLoading={isLoading}
      isError={isError}
      value={value}
      onChange={onChange}
      name={name}
      disabled={disabled}
    />
  )
}
