import { useState } from 'react'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import {
  getAccount,
  getOrgMembers,
  listRoles,
  removeUser,
  resendOrgInvite,
  revokeOrgInvite,
  updateAccountRole,
} from '@/lib'
import type { TOrgMember, TRoleInfo } from '@/types/ctl-api.types'
import { useListQueryState } from '../../../hooks/use-list-query-state'
import { useOrgAdmin } from '../../../hooks/use-org-admin'
import { useSurfaces } from '../../../hooks/use-surfaces'
import { useToast } from '../../../hooks/use-toast'
import { useOrg } from '../../../providers/org-provider'
import { commaSetQueryParameter } from '../../../utils/list-query'
import { ChangeRoleModal } from '../ChangeRoleModal'
import { RemoveMemberModal } from '../RemoveMemberModal'
import { ResendInviteModal } from '../ResendInviteModal'
import { RevokeInviteModal } from '../RevokeInviteModal'
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

const ChangeRoleModalContainer = ({
  member,
  orgId,
  roles,
}: {
  member: TOrgMember
  orgId: string
  roles: TRoleInfo[]
}) => {
  const queryClient = useQueryClient()
  const { closeTopSurface } = useSurfaces()
  const { addToast } = useToast()
  const mutation = useMutation({
    mutationFn: (roleType: string) =>
      updateAccountRole({
        orgId,
        accountId: member?.account_id ?? '',
        body: { role_type: roleType },
      }),
    onSuccess: (_account, roleType) => {
      void queryClient.invalidateQueries({ queryKey: ['org-members', orgId] })
      closeTopSurface()
      const title =
        roles.find((role) => role.role_type === roleType)?.title ?? roleType
      addToast({
        heading: 'Role updated',
        description: `${member?.email ?? 'Team member'} now has the ${title} role.`,
        theme: 'success',
      })
    },
  })

  return (
    <ChangeRoleModal
      email={member?.email ?? ''}
      roles={roles}
      currentRole={member?.role_type}
      onSubmit={(roleType) => mutation.mutate(roleType)}
      pending={mutation.isPending}
      error={mutation.error}
    />
  )
}

const RemoveMemberModalContainer = ({
  member,
  orgId,
}: {
  member: TOrgMember
  orgId: string
}) => {
  const [confirmation, setConfirmation] = useState('')
  const queryClient = useQueryClient()
  const { closeTopSurface } = useSurfaces()
  const { addToast } = useToast()
  const mutation = useMutation({
    mutationFn: () =>
      removeUser({
        orgId,
        body: { user_id: member?.account_id ?? '' },
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['org-members', orgId] })
      closeTopSurface()
      addToast({
        heading: 'Team member removed',
        description: `${member?.email ?? 'The team member'} no longer has access to this org.`,
        theme: 'success',
      })
    },
  })

  return (
    <RemoveMemberModal
      email={member?.email ?? ''}
      confirmation={confirmation}
      onConfirmationChange={setConfirmation}
      onSubmit={() => mutation.mutate()}
      pending={mutation.isPending}
      error={mutation.error}
    />
  )
}

const ResendInviteModalContainer = ({
  member,
  orgId,
}: {
  member: TOrgMember
  orgId: string
}) => {
  const queryClient = useQueryClient()
  const { closeTopSurface } = useSurfaces()
  const { addToast } = useToast()
  const mutation = useMutation({
    mutationFn: () =>
      resendOrgInvite({
        orgId,
        inviteId: member?.invite_id ?? '',
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['org-members', orgId] })
      closeTopSurface()
      addToast({
        heading: 'Invite sent',
        description: `Sent another invitation to ${member?.email ?? 'the team member'}.`,
        theme: 'success',
      })
    },
  })

  return (
    <ResendInviteModal
      email={member?.email ?? ''}
      onSubmit={() => mutation.mutate()}
      pending={mutation.isPending}
      error={mutation.error}
    />
  )
}

const RevokeInviteModalContainer = ({
  member,
  orgId,
}: {
  member: TOrgMember
  orgId: string
}) => {
  const queryClient = useQueryClient()
  const { closeTopSurface } = useSurfaces()
  const { addToast } = useToast()
  const mutation = useMutation({
    mutationFn: () =>
      revokeOrgInvite({
        orgId,
        inviteId: member?.invite_id ?? '',
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['org-members', orgId] })
      closeTopSurface()
      addToast({
        heading: 'Invite revoked',
        description: `${member?.email ?? 'The team member'} can no longer accept this invitation.`,
        theme: 'success',
      })
    },
  })

  return (
    <RevokeInviteModal
      email={member?.email ?? ''}
      onSubmit={() => mutation.mutate()}
      pending={mutation.isPending}
      error={mutation.error}
    />
  )
}

export const TeamTableContainer = () => {
  const { orgId } = useOrg()
  const admin = useOrgAdmin()
  const { openModal } = useSurfaces()
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
      onChangeRole={(member) =>
        openModal(
          <ChangeRoleModalContainer
            member={member}
            orgId={orgId ?? ''}
            roles={roles}
          />
        )
      }
      onRemove={(member) =>
        openModal(
          <RemoveMemberModalContainer member={member} orgId={orgId ?? ''} />
        )
      }
      onResend={(member) =>
        openModal(
          <ResendInviteModalContainer member={member} orgId={orgId ?? ''} />
        )
      }
      onRevoke={(member) =>
        openModal(
          <RevokeInviteModalContainer member={member} orgId={orgId ?? ''} />
        )
      }
      loading={isLoading}
      fetching={isPlaceholderData}
      error={error}
    />
  )
}
