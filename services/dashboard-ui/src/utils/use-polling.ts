'use client'

import { useEffect, useRef, useState } from 'react'
import type { TQuery, TQueryError } from './query-data'

type IUsePolling<T> = {
  dependencies?: Array<string> // extra dependencies for re-polling
  headers?: Record<string, string>
  initData?: T
  path: string
  pollInterval?: number
  shouldPoll?: boolean
}

export function usePolling<T = any>({
  dependencies = [],
  headers,
  initData = null,
  path,
  pollInterval = 5000,
  shouldPoll = false,
}: IUsePolling<T>) {
  const [data, setData] = useState<T | null>(initData)
  const [error, setError] = useState<TQueryError>(null)
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const intervalRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    if (!shouldPoll) return

    const poll = () => {
      setIsLoading(true)
      fetch(path, {
        headers,
      })
        .then((r) =>
          r.json().then((res: TQuery<T>) => {
            setIsLoading(false)
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
    }

    poll() // Fetch immediately
    intervalRef.current = setInterval(poll, pollInterval)

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [path, shouldPoll, pollInterval, initData, ...dependencies])

  return { data, error, isLoading }
}
