import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getAvailableRoles } from '@/lib'
import type { TOperationType, TPrincipalType, TWorkflowType } from '@/types'
import { RoleSelector } from './RoleSelector'

interface IRoleSelectorContainer {
  installId: string
  operationType?: TOperationType
  principalType?: TPrincipalType
  principalId?: string
  workflowType?: TWorkflowType
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
  workflowType,
  value,
  onChange,
  name,
  disabled,
}: IRoleSelectorContainer) => {
  const { org } = useOrg()

  const { data, isLoading, isError } = useQuery({
    queryKey: [
      'available-roles',
      org?.id,
      installId,
      operationType,
      principalType,
      principalId,
      workflowType,
    ],
    queryFn: () =>
      getAvailableRoles({
        installId,
        operationType,
        principalType,
        principalId,
        workflowType,
        orgId: org!.id,
      }),
    enabled: !!installId && !!org?.id,
  })

  const roles = data?.roles ?? []

  return (
    <RoleSelector
      roles={roles as any}
      isLoading={isLoading}
      isError={isError}
      isWorkflowDefault={!!workflowType && !principalType}
      value={value}
      onChange={onChange}
      name={name}
      disabled={disabled}
    />
  )
}
