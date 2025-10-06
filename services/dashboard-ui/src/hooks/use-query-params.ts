import { useMemo } from 'react'
import { buildQueryParams } from '@/utils/build-query-params'

/**
 * Returns a query string for any params, e.g. ?offset=10&limit=20&q=search
 * @param params - any query params object
 */
export function useQueryParams(params: Record<string, any>) {
  return useMemo(() => buildQueryParams(params), [params])
}
