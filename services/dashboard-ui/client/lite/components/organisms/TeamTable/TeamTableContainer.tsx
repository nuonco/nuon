import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getAccount, getOrgMembers, listRoles } from '@/lib'
import type { TOrgMember } from '@/types/ctl-api.types'
import { useListQueryState } from '../../../hooks/use-list-query-state'
import { useOrgAdmin } from '../../../hooks/use-org-admin'
import { useOrg } from '../../../providers/org-provider'
import { commaSetQueryParameter } from '../../../utils/list-query'
import { TeamTable, type ITeamFilter } from './TeamTable'

const PAGE_SIZE = 20
const TEAM_FILTERS = {
  status: commaSetQueryParameter('status'),
  roles: commaSetQueryParameter('role_type'),
}

const filterControl = ({
  label,
  options,
  constrained,
  onChange,
}: {
  label: string
  options: { value: string; label: string; description?: string }[]
  constrained: Set<string>
  onChange: (value: Set<string>) => void
}): ITeamFilter => ({
  label,
  options,
  selected: constrained,
  constrained: constrained.size > 0,
  onToggle: (value) => {
    const next = new Set(constrained)
    if (next.has(value)) next.delete(value)
    else next.add(value)
    onChange(next)
  },
  onIsolate: (value) =>
    onChange(
      constrained.size === 1 && constrained.has(value)
        ? new Set()
        : new Set([value])
    ),
  onReset: () => onChange(new Set()),
})

const ignoreMember = (_member: TOrgMember) => {}

export const TeamTableContainer = () => {
  const { orgId } = useOrg()
  const admin = useOrgAdmin()
  const list = useListQueryState({
    pageSize: PAGE_SIZE,
    filters: TEAM_FILTERS,
  })

  const {
    data: result,
    isLoading,
    isPlaceholderData,
    error,
  } = useQuery({
    queryKey: ['org-members', orgId, ...list.queryKey],
    queryFn: () =>
      getOrgMembers({
        orgId: orgId!,
        q: list.search || undefined,
        status: [...list.filters.status].join(',') || undefined,
        role_type: [...list.filters.roles].join(',') || undefined,
        offset: list.offset,
        limit: list.pageSize,
      }),
    enabled: !!orgId,
    placeholderData: keepPreviousData,
    refetchInterval: 20_000,
  })

  const { data: roles = [] } = useQuery({
    queryKey: ['roles', orgId, 'team'],
    queryFn: () => listRoles({ orgId: orgId!, context: 'team' }),
    enabled: !!orgId,
    staleTime: 60_000,
  })

  const { data: account } = useQuery({
    queryKey: ['account'],
    queryFn: getAccount,
    staleTime: 60_000,
  })

  const statusFilter = filterControl({
    label: 'Status',
    options: [
      { value: 'active', label: 'Active' },
      { value: 'invited', label: 'Invited' },
    ],
    constrained: list.filters.status,
    onChange: (value) => list.setFilter('status', value),
  })

  const roleFilter = roles.length
    ? filterControl({
        label: 'Role',
        options: roles.map((role) => ({
          value: role.role_type,
          label: role.title || role.role_type,
          description: role.description,
        })),
        constrained: list.filters.roles,
        onChange: (value) => list.setFilter('roles', value),
      })
    : undefined

  return (
    <TeamTable
      members={result?.data ?? []}
      roles={roles}
      currentAccountId={account?.id}
      admin={admin}
      search={list.search}
      onSearchChange={list.setSearch}
      offset={list.offset}
      pageSize={list.pageSize}
      hasNext={result?.pagination?.hasNext ?? false}
      onOffsetChange={list.setOffset}
      statusFilter={statusFilter}
      roleFilter={roleFilter}
      onChangeRole={ignoreMember}
      onRemove={ignoreMember}
      onResend={ignoreMember}
      onRevoke={ignoreMember}
      loading={isLoading}
      fetching={isPlaceholderData}
      error={error}
    />
  )
}
