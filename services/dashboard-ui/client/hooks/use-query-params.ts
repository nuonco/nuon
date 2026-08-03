import { useMemo } from 'react'
import { buildQueryParams } from '@/utils/build-query-params'

export function useQueryParams(params: Record<string, any>) {
  return useMemo(() => buildQueryParams(params), [params])
}
