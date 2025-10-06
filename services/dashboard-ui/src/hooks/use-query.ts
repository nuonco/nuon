'use client'

import { useEffect, useState } from 'react'
import type { TAPIResponse, TAPIError } from '@/types'

interface IUseQuery<T> {
  path: string
  headers?: Record<string, string>
  dependencies?: Array<any> // dependencies for re-fetching
  initData?: T | null
  enabled?: boolean // If false, don't fetch
}

export function useQuery<T = any>({
  path,
  headers,
  dependencies = [],
  initData = null,
  enabled = true,
}: IUseQuery<T>) {
  const [data, setData] = useState<T | null>(initData)
  const [error, setError] = useState<TAPIError | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(true)
  const [responseHeaders, setResponseHeaders] = useState<Record<
    string,
    string
  > | null>(null)
  const [status, setStatus] = useState<number | null>(null)

  useEffect(() => {
    if (!enabled) return

    setIsLoading(true)
    fetch(path, { headers })
      .then((r) =>
        r.json().then((res: TAPIResponse<T>) => {
          setIsLoading(false)
          setStatus(res?.status || 500)
          setResponseHeaders(res?.headers)
          if (res?.error) {
            setError(res?.error)
          } else {
            setError(null)
            setData(res.data)
          }
        })
      )
      .catch((err) => {
        setIsLoading(false)
        setError(err)
        setResponseHeaders(null)
      })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [path, enabled, ...dependencies])

  return { data, error, isLoading, headers: responseHeaders, status }
}
