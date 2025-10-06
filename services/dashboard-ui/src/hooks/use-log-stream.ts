'use client'

import { useContext } from 'react'
import { LogStreamContext } from '@/providers/log-stream-provider'

export function useLogStream() {
  const ctx = useContext(LogStreamContext)
  if (!ctx) {
    throw new Error('useLogStream must be used within an LogStreamProvider')
  }
  return ctx
}
