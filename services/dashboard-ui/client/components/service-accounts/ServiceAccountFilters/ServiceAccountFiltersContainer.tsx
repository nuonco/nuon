import { useEffect, useRef, useState } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useServiceAccountFilters } from '@/hooks/use-service-account-filters'
import { useToast } from '@/hooks/use-toast'
import { listServiceAccountPurposes } from '@/lib'
import type { TAPIError } from '@/types'
import { ServiceAccountFilters } from './ServiceAccountFilters'

const SEARCH_DEBOUNCE_MS = 300

export const ServiceAccountFiltersContainer = () => {
  const { org } = useOrg()
  const { addToast } = useToast()
  const orgId = org?.id
  const {
    management,
    purpose,
    q,
    needsNormalization,
    setFilters,
    normalizeFilters,
  } = useServiceAccountFilters()
  const [search, setSearch] = useState(q)
  const committedSearch = useRef(q)

  useEffect(() => {
    if (needsNormalization) normalizeFilters()
  }, [needsNormalization, normalizeFilters])

  useEffect(() => {
    if (q === committedSearch.current) return
    committedSearch.current = q
    setSearch(q)
  }, [q])

  useEffect(() => {
    const next = search.trim()
    if (next === committedSearch.current) return
    const timeout = setTimeout(() => {
      committedSearch.current = next
      setFilters({ q: next }, { replace: true })
    }, SEARCH_DEBOUNCE_MS)
    return () => clearTimeout(timeout)
  }, [search, setFilters])

  const { data: purposes, error, errorUpdatedAt } = useQuery({
    queryKey: ['service-accounts', orgId, 'purposes'],
    queryFn: () => listServiceAccountPurposes({ orgId: orgId! }),
    enabled: !!orgId && management !== 'user',
    placeholderData: keepPreviousData,
  })

  const lastPurposeError = useRef<string>()
  useEffect(() => {
    if (!orgId || management === 'user' || !error) return
    const errorKey = `${orgId}:${errorUpdatedAt}`
    if (lastPurposeError.current === errorKey) return
    lastPurposeError.current = errorKey
    const apiError = error as TAPIError
    addToast(
      <Toast heading="Purpose loading failed" theme="error">
        <Text>
          {apiError.description ||
            apiError.error ||
            'Unable to load purpose filters. Refresh the page to try again.'}
        </Text>
      </Toast>
    )
  }, [orgId, management, error, errorUpdatedAt, addToast])

  return (
    <ServiceAccountFilters
      management={management}
      purpose={purpose}
      purposes={purposes ?? []}
      search={search}
      onManagementChange={(value) => setFilters({ management: value })}
      onPurposeChange={(value) => setFilters({ purpose: value })}
      onSearchChange={setSearch}
    />
  )
}
