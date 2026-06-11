import { useState, useRef, useEffect, useCallback } from 'react'
import { useToast } from '@/hooks/use-toast'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import {
  ensureActivityTracking,
  isRecentlyActive,
  onNextActivity,
} from '@/lib/user-activity'

const RECENT_ACTIVITY_WINDOW_MS = 5 * 60_000

type SSEEventHandler = (event: MessageEvent) => void

interface UseResourceSSEOptions {
  url: string | undefined
  enabled: boolean
  onMessage?: SSEEventHandler
  listeners?: Record<string, SSEEventHandler>
}

export function useResourceSSE({ url, enabled, onMessage, listeners }: UseResourceSSEOptions) {
  const { addToast } = useToast()
  const [connected, setConnected] = useState(false)
  const [suspended, setSuspended] = useState(false)
  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectAttemptRef = useRef(0)
  const finishedRef = useRef(false)
  const expiredRef = useRef(false)
  const resumeCleanupRef = useRef<(() => void) | null>(null)
  const onMessageRef = useRef(onMessage)
  onMessageRef.current = onMessage
  const listenersRef = useRef(listeners)
  listenersRef.current = listeners

  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
    if (resumeCleanupRef.current) {
      resumeCleanupRef.current()
      resumeCleanupRef.current = null
    }
    setSuspended(false)
    setConnected(false)
  }, [])

  const connect = useCallback(() => {
    if (!url || eventSourceRef.current) return

    if (resumeCleanupRef.current) {
      resumeCleanupRef.current()
      resumeCleanupRef.current = null
    }
    setSuspended(false)
    expiredRef.current = false

    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    if (onMessageRef.current) {
      eventSource.onmessage = (event) => onMessageRef.current?.(event)
    }

    if (listenersRef.current) {
      for (const event of Object.keys(listenersRef.current)) {
        eventSource.addEventListener(event, (e: MessageEvent) =>
          listenersRef.current?.[event]?.(e)
        )
      }
    }

    eventSource.addEventListener('finished', () => {
      finishedRef.current = true
    })

    eventSource.addEventListener('expired', () => {
      expiredRef.current = true
    })

    eventSource.addEventListener('fetch-error', (event: MessageEvent) => {
      try {
        const errorData = JSON.parse(event.data)
        addToast(
          <Toast heading="Refresh failed" theme="warn">
            <Text>{errorData?.error ?? 'Connection issue'}</Text>
          </Toast>
        )
      } catch {
        // non-JSON error event
      }
    })

    eventSource.onerror = () => {
      eventSource.close()
      eventSourceRef.current = null
      setConnected(false)

      // The server closes finished resources for good after a grace period;
      // fallback polling takes over instead of reconnecting forever.
      if (finishedRef.current) return

      // The server expires streams after a max lifetime. If the user hasn't
      // interacted recently, suspend until their next interaction instead of
      // reconnecting — an unattended visible tab shouldn't poll at full rate.
      if (expiredRef.current && !isRecentlyActive(RECENT_ACTIVITY_WINDOW_MS)) {
        setSuspended(true)
        resumeCleanupRef.current = onNextActivity(() => {
          resumeCleanupRef.current = null
          connect()
        })
        return
      }

      const backoffDelay = Math.min(1000 * Math.pow(2, reconnectAttemptRef.current), 30000)
      reconnectAttemptRef.current += 1

      reconnectTimeoutRef.current = setTimeout(() => {
        connect()
      }, backoffDelay)
    }

    eventSource.onopen = () => {
      setConnected(true)
      reconnectAttemptRef.current = 0
    }
  }, [url])

  useEffect(() => {
    ensureActivityTracking()
    finishedRef.current = false
    if (enabled && url && document.visibilityState !== 'hidden') {
      connect()
    }
    return () => disconnect()
  }, [enabled, url, connect, disconnect])

  // Hidden tabs disconnect so the BFF stops polling ctl-api for them. On
  // return, the focus refetch covers the gap and the fresh stream sends a
  // full snapshot on its first tick.
  useEffect(() => {
    const onVisibilityChange = () => {
      if (document.visibilityState === 'hidden') {
        disconnect()
      } else if (enabled && url && !finishedRef.current) {
        reconnectAttemptRef.current = 0
        connect()
      }
    }
    document.addEventListener('visibilitychange', onVisibilityChange)
    return () => document.removeEventListener('visibilitychange', onVisibilityChange)
  }, [enabled, url, connect, disconnect])

  return { connected, suspended, disconnect }
}
