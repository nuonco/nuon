import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getRole, listRoles } from '@/lib'
import type { TRoleContext } from '@/types'

// roleAssignmentID mirrors Role.AssignmentIdentifier in
// services/ctl-api/internal/app/role.go: managed roles are addressed by type,
// custom roles by id. Every custom role shares the "custom" role_type, so the
// type alone is ambiguous and the API rejects the bare string.
export function roleAssignmentID(
  role: { id?: string; role_type?: string } | undefined | null
): string {
  if (!role) return ''
  if (role.role_type === 'custom' && role.id) return role.id
  return role.role_type ?? ''
}

export function useRoles(context?: TRoleContext) {
  const { org } = useOrg()

  return useQuery({
    queryKey: ['roles', org?.id, context ?? 'all'],
    queryFn: () => listRoles({ orgId: org!.id, context }),
    enabled: !!org?.id,
    staleTime: 5 * 60 * 1000,
  })
}

export function useRole(roleId: string | undefined) {
  const { org } = useOrg()

  return useQuery({
    queryKey: ['role', org?.id, roleId],
    queryFn: () => getRole({ orgId: org!.id, roleId: roleId! }),
    enabled: !!org?.id && !!roleId,
  })
}

export function useRoleOptions(context: TRoleContext) {
  const { data: roles, isLoading } = useRoles(context)

  const roleOptions = useMemo(
    () =>
      (roles ?? []).map((role) => ({
        value: roleAssignmentID(role),
        label: role.title || role.role_type,
        description: role.description,
      })),
    [roles]
  )

  return { roleOptions, isLoading }
}

export function useRoleTitles() {
  const { data: roles } = useRoles()

  return useMemo(() => {
    // Keyed under both identifiers so a title resolves whichever one a row
    // stores: assignments made through a picker carry the assignment id,
    // while older rows and managed roles carry the role type.
    const lookup = (roles ?? []).reduce<Record<string, string>>((acc, role) => {
      const title = role.title || role.role_type
      if (role.id) acc[role.id] = title
      // Custom roles all share the "custom" role_type, so keying them by type
      // would have them overwrite each other under one entry.
      if (role.role_type !== 'custom') acc[role.role_type] = title
      return acc
    }, {})

    return (roleType: string | undefined) => {
      if (!roleType) return '—'
      return lookup[roleType] ?? roleType
    }
  }, [roles])
}
