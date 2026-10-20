import { useRef, useState } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { getOrgMembers } from '@/lib'
import { ActorFilter } from './ActorFilter'

const SEARCH_DEBOUNCE_MS = 300
const MEMBER_LIMIT = 20

export interface IActorFilterContainer {
  orgId: string
  selected?: string
  onChange: (accountId?: string) => void
}

export const ActorFilterContainer = ({
  orgId,
  selected,
  onChange,
}: IActorFilterContainer) => {
  const { user } = useAuth()
  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const handleQueryChange = (value: string) => {
    setQuery(value)
    if (debounceTimer.current) clearTimeout(debounceTimer.current)
    debounceTimer.current = setTimeout(
      () => setDebouncedQuery(value),
      SEARCH_DEBOUNCE_MS
    )
  }

  const { data, isLoading } = useQuery({
    queryKey: ['org-members-for-actor-filter', orgId, debouncedQuery],
    queryFn: () =>
      getOrgMembers({
        orgId,
        q: debouncedQuery || undefined,
        status: 'active',
        limit: MEMBER_LIMIT,
      }),
    enabled: !!orgId,
    placeholderData: keepPreviousData,
  })

  const members = (data?.data ?? [])
    .filter(
      (member) => !!member.account_id && member.account_id !== user?.sub
    )
    .map((member) => ({
      accountId: member.account_id as string,
      email: member.email ?? '',
    }))

  return (
    <ActorFilter
      id="activity-filter-actor"
      currentUser={
        user?.sub ? { accountId: user.sub, email: user.email ?? '' } : undefined
      }
      members={members}
      isLoading={isLoading}
      query={query}
      onQueryChange={handleQueryChange}
      selected={selected}
      onChange={onChange}
    />
  )
}
