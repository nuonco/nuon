import { useCallback } from 'react'
import { useSearchParams } from 'react-router'
import type { TServiceAccountManagement } from '@/types'

export const SERVICE_ACCOUNT_MANAGEMENT_PARAM = 'management'
export const SERVICE_ACCOUNT_PURPOSE_PARAM = 'purpose'
export const SERVICE_ACCOUNT_SEARCH_PARAM = 'q'
const LEGACY_PARAMS = ['runners', 'stacks'] as const

const MANAGEMENT_VALUES: TServiceAccountManagement[] = ['user', 'system', 'all']

export type TServiceAccountFilterUpdate = Partial<{
  management: TServiceAccountManagement
  purpose: string
  q: string
}>

const hasLegacyParams = (params: URLSearchParams) =>
  LEGACY_PARAMS.some((param) => params.get(param) === 'true')

const isManagement = (
  value: string | null
): value is TServiceAccountManagement =>
  MANAGEMENT_VALUES.includes(value as TServiceAccountManagement)

export function parseServiceAccountManagement(
  params: URLSearchParams
): TServiceAccountManagement {
  const value = params.get(SERVICE_ACCOUNT_MANAGEMENT_PARAM)
  if (isManagement(value)) return value
  return hasLegacyParams(params) ? 'all' : 'user'
}

const setOrDelete = (params: URLSearchParams, key: string, value?: string) => {
  if (value) {
    params.set(key, value)
  } else {
    params.delete(key)
  }
}

export function applyServiceAccountFilters(
  prev: URLSearchParams,
  updates: TServiceAccountFilterUpdate = {}
): URLSearchParams {
  const params = new URLSearchParams(prev)
  const management = updates.management ?? parseServiceAccountManagement(params)
  LEGACY_PARAMS.forEach((param) => params.delete(param))
  setOrDelete(
    params,
    SERVICE_ACCOUNT_MANAGEMENT_PARAM,
    management === 'user' ? undefined : management
  )
  if ('purpose' in updates) {
    setOrDelete(params, SERVICE_ACCOUNT_PURPOSE_PARAM, updates.purpose)
  }
  if (management === 'user') {
    params.delete(SERVICE_ACCOUNT_PURPOSE_PARAM)
  }
  if ('q' in updates) {
    setOrDelete(params, SERVICE_ACCOUNT_SEARCH_PARAM, updates.q?.trim())
  }
  params.delete('offset')
  return params
}

export function useServiceAccountFilters() {
  const [searchParams, setSearchParams] = useSearchParams()

  const management = parseServiceAccountManagement(searchParams)
  const purpose =
    management === 'user'
      ? ''
      : (searchParams.get(SERVICE_ACCOUNT_PURPOSE_PARAM) ?? '')
  const q = searchParams.get(SERVICE_ACCOUNT_SEARCH_PARAM)?.trim() ?? ''
  const offset = Math.max(0, Number(searchParams.get('offset')) || 0)
  const needsNormalization =
    LEGACY_PARAMS.some((param) => searchParams.has(param)) ||
    (management === 'user' && searchParams.has(SERVICE_ACCOUNT_PURPOSE_PARAM))

  const setFilters = useCallback(
    (
      updates: TServiceAccountFilterUpdate,
      options: { replace?: boolean } = {}
    ) => {
      setSearchParams((prev) => applyServiceAccountFilters(prev, updates), {
        replace: options.replace,
      })
    },
    [setSearchParams]
  )

  const normalizeFilters = useCallback(() => {
    setSearchParams(
      (prev) => {
        const params = applyServiceAccountFilters(prev)
        const offset = prev.get('offset')
        if (offset) params.set('offset', offset)
        return params
      },
      { replace: true }
    )
  }, [setSearchParams])

  return {
    management,
    purpose,
    q,
    offset,
    needsNormalization,
    setFilters,
    normalizeFilters,
  }
}
