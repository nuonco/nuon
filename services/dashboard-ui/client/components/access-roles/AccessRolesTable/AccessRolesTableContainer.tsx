import { useCallback, useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useRoles } from '@/hooks/use-roles'
import { getApps, getInstalls } from '@/lib'
import { AccessRolesTable } from './AccessRolesTable'

const RESOURCE_LIMIT = 100

export const AccessRolesTableContainer = () => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { data: roles, isLoading } = useRoles()
  const query = (searchParams.get('q') ?? '').toLowerCase()

  const { data: apps } = useQuery({
    queryKey: ['permission-entry-apps', org?.id],
    queryFn: () => getApps({ orgId: org!.id, limit: RESOURCE_LIMIT }),
    enabled: !!org?.id,
    staleTime: 5 * 60 * 1000,
  })

  const { data: installs } = useQuery({
    queryKey: ['permission-entry-installs', org?.id],
    queryFn: () => getInstalls({ orgId: org!.id, limit: RESOURCE_LIMIT }),
    enabled: !!org?.id,
    staleTime: 5 * 60 * 1000,
  })

  const names = useMemo(() => {
    const lookup: Record<string, string> = {}
    for (const item of [...(apps?.data ?? []), ...(installs?.data ?? [])]) {
      if (item?.id && item?.name) lookup[item.id] = item.name
    }
    return lookup
  }, [apps?.data, installs?.data])

  const nameFor = useCallback((id: string) => names[id], [names])

  const filtered = (roles ?? []).filter((role) =>
    query
      ? `${role.title ?? ''} ${role.description ?? ''}`
          .toLowerCase()
          .includes(query)
      : true
  )

  return (
    <AccessRolesTable
      data={filtered}
      isLoading={isLoading}
      nameFor={nameFor}
    />
  )
}
