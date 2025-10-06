'use client'

import { useEffect, useRef, useState } from 'react'
import { POLLING_TIMEOUT } from '@/configs/api'
import type { TAPIResponse, TAPIError } from '@/types'

export interface IPollingProps {
  pollInterval?: number
  shouldPoll?: boolean
}

interface IUsePolling<T> {
  dependencies?: Array<any> // extra dependencies for re-polling
  headers?: Record<string, string>
  initData?: T | null
  initIsLoading?: boolean
  path: string
  pollInterval?: number
  shouldPoll?: boolean
}

export function usePolling<T = any>({
  dependencies = [],
  headers,
  initData = null,
  initIsLoading = false,
  path,
  pollInterval = POLLING_TIMEOUT,
  shouldPoll = false,
}: IUsePolling<T>) {
  const [data, setData] = useState<T | null>(initData)
  const [error, setError] = useState<TAPIError | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(initIsLoading)
  const [responseHeaders, setResponseHeaders] = useState<Record<
    string,
    string
  > | null>(null)
  const [status, setStatus] = useState<number | null>(null)

  const intervalRef = useRef<NodeJS.Timeout | null>(null)

  const stopPolling = () => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
  }

  useEffect(() => {
    if (!shouldPoll) return

    setIsLoading(true)
    const poll = () => {
      fetch(path, {
        headers,
      })
        .then((response) => {
          return response.json().then((res: TAPIResponse<T>) => {
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
        })
        .catch((err) => {
          setIsLoading(false)
          setError(err)
          setResponseHeaders(null)
        })
    }

    poll() // Fetch immediately
    intervalRef.current = setInterval(poll, pollInterval)

    return () => {
      stopPolling()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [path, shouldPoll, pollInterval, ...dependencies])

  useEffect(() => {
    setData(initData)
    setError(null)
    setIsLoading(false)
    setResponseHeaders(null)
  }, [initData])

  return {
    data,
    error,
    isLoading,
    headers: responseHeaders,
    status,
    stopPolling,
  }
}
