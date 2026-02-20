'use client'

import { useCallback, useState } from 'react'
import type { TAPIError, TAPIResponse } from '@/types'

/**
 * useMutation — REST-based mutation hook for Go BFF mode.
 * POSTs JSON to the given endpoint and returns the same shape as useServerAction.
 */
export function useMutation<TArgs, TData>(endpoint: string) {
  const [data, setData] = useState<TData | null>(null)
  const [error, setError] = useState<TAPIError | null>(null)
  const [headers, setHeaders] = useState<Record<string, string> | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const [status, setStatus] = useState<number | null>(null)

  const execute = useCallback(
    async (args: TArgs): Promise<TAPIResponse<TData>> => {
      setIsLoading(true)
      setError(null)
      setStatus(null)
      setHeaders(null)

      try {
        const res = await fetch(endpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(args),
        })

        const json: TAPIResponse<TData> = await res.json()

        setData(json.data)
        setError(json.error)
        setStatus(json.status)
        setHeaders(json.headers)
        return json
      } catch (err: any) {
        const errorResponse: TAPIResponse<TData> = {
          data: null,
          error: err,
          status: null as any,
          headers: null as any,
        }
        setData(null)
        setError(err)
        setStatus(null)
        setHeaders(null)
        return errorResponse
      } finally {
        setIsLoading(false)
      }
    },
    [endpoint],
  )

  return { data, error, status, headers, isLoading, execute }
}
