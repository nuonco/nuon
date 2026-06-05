import { useState, useRef, useEffect, useCallback } from 'react'
import { useToast } from '@/hooks/use-toast'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'

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
  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectAttemptRef = useRef(0)
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
    setConnected(false)
  }, [])

  const connect = useCallback(() => {
    if (!url || eventSourceRef.current) return

    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    if (onMessageRef.current) {
      const handler = onMessageRef.current
      eventSource.onmessage = handler
    }

    if (listenersRef.current) {
      for (const [event, handler] of Object.entries(listenersRef.current)) {
        eventSource.addEventListener(event, handler)
      }
    }

    eventSource.addEventListener('finished', () => {})

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
    if (enabled && url) {
      connect()
    }
    return () => disconnect()
  }, [enabled, url, connect, disconnect])

  return { connected, disconnect }
}
