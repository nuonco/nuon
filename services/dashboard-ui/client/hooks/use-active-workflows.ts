import { useContext } from 'react'
import { ActiveWorkflowsContext } from '@/providers/active-workflows-provider'

export function useActiveWorkflows() {
  const ctx = useContext(ActiveWorkflowsContext)
  if (!ctx) {
    throw new Error('useActiveWorkflows must be used within an ActiveWorkflowsProvider')
  }
  return ctx
}
