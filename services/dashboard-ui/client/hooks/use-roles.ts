import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { listRoles } from '@/lib'
import type { TRoleContext } from '@/types'

export function useRoles(context?: TRoleContext) {
  const { org } = useOrg()

  return useQuery({
    queryKey: ['roles', org?.id, context ?? 'all'],
    queryFn: () => listRoles({ orgId: org!.id, context }),
    enabled: !!org?.id,
    staleTime: 5 * 60 * 1000,
  })
}

export function useRoleOptions(context: TRoleContext) {
  const { data: roles, isLoading } = useRoles(context)

  const roleOptions = useMemo(
    () =>
      (roles ?? []).map((role) => ({
        value: role.role_type,
        label: role.title || role.role_type,
      })),
    [roles]
  )

  return { roleOptions, isLoading }
}

export function useRoleTitles() {
  const { data: roles } = useRoles()

  return useMemo(() => {
    const lookup = (roles ?? []).reduce<Record<string, string>>((acc, role) => {
      acc[role.role_type] = role.title || role.role_type
      return acc
    }, {})

    return (roleType: string | undefined) => {
      if (!roleType) return '—'
      return lookup[roleType] ?? roleType
    }
  }, [roles])
}
