import { useContext } from 'react'
import { WorkflowContext } from '@/providers/WorkflowProvider'

export const useWorkflow = () => {
  const context = useContext(WorkflowContext)
  if (context === undefined) {
    throw new Error('useWorkflow must be used within a WorkflowProvider')
  }
  return context
}