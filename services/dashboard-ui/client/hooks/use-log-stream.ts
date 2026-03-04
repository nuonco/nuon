import { useContext } from 'react'
import { LogStreamContext } from '@/providers/log-stream-provider'
import type { TLogStream } from '@/types'

export function useLogStream(): { logStream: TLogStream; refresh: () => void } {
  const ctx = useContext(LogStreamContext)
  if (!ctx) {
    throw new Error('useLogStream must be used within an LogStreamProvider')
  }
  return ctx
}
