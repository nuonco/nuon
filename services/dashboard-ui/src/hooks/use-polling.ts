'use client'

import { useEffect, useRef, useState } from 'react'
import { POLLING_TIMEOUT } from '@/configs/api'
import type { TAPIResponse, TAPIError } from '@/types'

export interface IPollingProps {
  pollInterval?: number
  shouldPoll?: boolean
}

interface IBackoffOptions {
  enabled?: boolean
  initialDelay?: number // ms
  maxDelay?: number // ms
  multiplier?: number
  jitter?: boolean
  maxRetries?: number // optional: stop backing off after this many failures (undefined = infinite)
  resetOnSuccess?: boolean
}

interface IUsePolling<T> {
  dependencies?: Array<any> // extra dependencies for re-polling
  headers?: Record<string, string>
  initData?: T | null
  initIsLoading?: boolean
  path: string
  pollInterval?: number
  shouldPoll?: boolean
  // Backoff-related optional fields (added but do not change existing fields)
  requestTimeout?: number
  backoff?: IBackoffOptions
}

export function usePolling<T = any>({
  dependencies = [],
  headers,
  initData = null,
  initIsLoading = false,
  path,
  pollInterval = POLLING_TIMEOUT,
  shouldPoll = false,
  // new options with sensible defaults
  requestTimeout = 10000,
  backoff = {
    enabled: true,
    initialDelay: 1000,
    maxDelay: 60000,
    multiplier: 2,
    jitter: true,
    maxRetries: undefined,
    resetOnSuccess: true,
  },
}: IUsePolling<T>) {
  const [data, setData] = useState<T | null>(initData)
  const [error, setError] = useState<TAPIError | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(initIsLoading)
  const [responseHeaders, setResponseHeaders] = useState<Record<
    string,
    string
  > | null>(null)
  const [status, setStatus] = useState<number | null>(null)

  // timer ref (will store window.setTimeout id)
  const intervalRef = useRef<NodeJS.Timeout | null>(null)
  // in-flight controller ref to prevent overlapping requests & to support abort
  const inFlightRef = useRef<AbortController | null>(null)
  // backoff tracking
  const currentDelayRef = useRef<number>(backoff?.initialDelay ?? 1000)
  const retryCountRef = useRef<number>(0)
  const mountedRef = useRef<boolean>(true)

  const stopPolling = () => {
    if (intervalRef.current) {
      try {
        clearTimeout(intervalRef.current as unknown as number)
      } catch {
        // noop
      }
      intervalRef.current = null
    }
    if (inFlightRef.current) {
      try {
        inFlightRef.current.abort()
      } catch {
        // noop
      }
      inFlightRef.current = null
    }
  }

  // compute next backoff delay (with optional jitter)
  const computeNextDelay = (base: number) => {
    const multiplier = backoff?.multiplier ?? 2
    const maxDelay = backoff?.maxDelay ?? 60000
    let next = Math.min(Math.floor(base * multiplier), maxDelay)
    if (backoff?.jitter) {
      // full jitter: random between base and next
      const min = Math.min(base, next)
      const max = Math.max(base, next)
      next = Math.floor(Math.random() * (max - min + 1) + min)
    }
    return next
  }

  useEffect(() => {
    mountedRef.current = true
    return () => {
      mountedRef.current = false
      stopPolling()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (!shouldPoll) {
      stopPolling()
      return
    }

    // helper to schedule the next poll via setTimeout
    const scheduleNext = (delay: number) => {
      if (!mountedRef.current) return
      // clear any existing timer
      if (intervalRef.current) {
        try {
          clearTimeout(intervalRef.current as unknown as number)
        } catch {
          // noop
        }
        intervalRef.current = null
      }
      intervalRef.current = setTimeout(() => {
        intervalRef.current = null
        void poll()
      }, delay) as unknown as NodeJS.Timeout
    }

    // fetch wrapper that supports timeout via AbortController
    const fetchWithTimeout = (input: RequestInfo, init?: RequestInit) => {
      const controller = new AbortController()
      const timeoutId = window.setTimeout(() => controller.abort(), requestTimeout)
      inFlightRef.current = controller
      return fetch(input, { ...init, signal: controller.signal }).finally(() => {
        clearTimeout(timeoutId)
        inFlightRef.current = null
      })
    }

    const poll = async () => {
      if (!shouldPoll || !mountedRef.current) return

      // prevent overlapping requests
      if (inFlightRef.current) {
        // schedule next after normal pollInterval (do not start a concurrent fetch)
        scheduleNext(pollInterval)
        return
      }

      setIsLoading(true)
      try {
        const response = await fetchWithTimeout(path, { headers })
        // attempt to parse JSON; if parse fails, treat as error
        const res = (await response.json().catch((e) => {
          throw e
        })) as TAPIResponse<T>

        if (!mountedRef.current) return

        // success
        setIsLoading(false)
        setStatus(res?.status || response.status || 200)
        setResponseHeaders(res?.headers ?? null)

        if (res?.error) {
          setError(res.error)
        } else {
          setError(null)
          setData(res.data)
        }

        // reset backoff state on success if configured
        if (backoff?.enabled && backoff.resetOnSuccess !== false) {
          currentDelayRef.current = backoff?.initialDelay ?? 1000
          retryCountRef.current = 0
        }

        // schedule next regular poll
        scheduleNext(pollInterval)
      } catch (err) {
        if (!mountedRef.current) return
        setIsLoading(false)
        setError(err as TAPIError)
        setResponseHeaders(null)

        if (backoff?.enabled) {
          retryCountRef.current += 1
          if (
            typeof backoff.maxRetries === 'number' &&
            retryCountRef.current > (backoff.maxRetries ?? 0)
          ) {
            // stop polling if maxRetries exceeded
            stopPolling()
            return
          }

          // compute next delay using exponential backoff
          const next = computeNextDelay(currentDelayRef.current)
          currentDelayRef.current = next
          scheduleNext(next)
        } else {
          // no backoff, just schedule normally
          scheduleNext(pollInterval)
        }
      }
    }

    // start immediately
    void poll()

    return () => {
      // cleanup timers and abort controllers when dependencies change/unmount
      if (intervalRef.current) {
        try {
          clearTimeout(intervalRef.current as unknown as number)
        } catch {
          // noop
        }
        intervalRef.current = null
      }
      if (inFlightRef.current) {
        try {
          inFlightRef.current.abort()
        } catch {
          // noop
        }
        inFlightRef.current = null
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [path, shouldPoll, pollInterval, requestTimeout, backoff?.enabled, backoff?.initialDelay, backoff?.maxDelay, backoff?.multiplier, backoff?.jitter, backoff?.maxRetries, backoff?.resetOnSuccess, ...dependencies])

  useEffect(() => {
    setData(initData)
    setError(null)
    setIsLoading(false)
    setResponseHeaders(null)
    // reset backoff trackers when initData changes
    currentDelayRef.current = backoff?.initialDelay ?? 1000
    retryCountRef.current = 0
  }, [initData, backoff?.initialDelay])

  return {
    data,
    error,
    isLoading,
    headers: responseHeaders,
    status,
    stopPolling,
  }
}
