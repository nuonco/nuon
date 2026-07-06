import { useEffect } from 'react'
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
  hideWhenEmpty?: boolean
  onAvailabilityChange?: (available: boolean) => void
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
  hideWhenEmpty,
  onAvailabilityChange,
}: IRoleSelectorContainer) => {
  const { org } = useOrg()

  const { data, isLoading, isError } = useQuery({
    queryKey: ['available-roles', org.id, installId, operationType, principalType, principalId],
    queryFn: () =>
      getAvailableRoles({ installId, operationType, principalType, principalId, orgId: org.id }),
    enabled: !!installId && !!org.id,
  })

  const roles = data?.roles ?? []

  useEffect(() => {
    if (!onAvailabilityChange) return
    onAvailabilityChange(isLoading || isError || roles.length > 0)
  }, [isLoading, isError, roles.length, onAvailabilityChange])

  return (
    <RoleSelector
      roles={roles as any}
      isLoading={isLoading}
      isError={isError}
      value={value}
      onChange={onChange}
      name={name}
      disabled={disabled}
      hideWhenEmpty={hideWhenEmpty}
    />
  )
}
