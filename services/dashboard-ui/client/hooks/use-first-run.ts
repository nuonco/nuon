import { useContext } from 'react'
import { FirstRunContext, type IFirstRunContext } from '@/providers/first-run-provider'

export function useFirstRun(): IFirstRunContext {
  const ctx = useContext(FirstRunContext)
  if (!ctx) {
    throw new Error('useFirstRun must be used within a FirstRunProvider')
  }
  return ctx
}
