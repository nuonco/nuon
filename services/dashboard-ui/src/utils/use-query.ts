'use client'

import { useEffect, useState } from 'react'
import type { TQuery, TQueryError } from './query-data'

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
  const [resHeaders, setResHeaders] = useState<Record<string, unknown> | null>(
    null
  )
  const [error, setError] = useState<TQueryError | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(true)

  useEffect(() => {
    if (!enabled) return

    setIsLoading(true)
    fetch(path, { headers })
      .then((r) =>
        r.json().then((res: TQuery<T>) => {
          setIsLoading(false)
          setResHeaders(res?.headers as unknown as Record<string, unknown>)
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
      })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [path, enabled, ...dependencies])

  return { data, error, headers: resHeaders, isLoading }
}
