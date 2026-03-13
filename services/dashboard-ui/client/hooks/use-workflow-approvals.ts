import { useContext } from 'react'
import { WorkflowApprovalsContext } from '@/providers/workflow-approvals-provider'

export function useWorkflowApprovals() {
  const ctx = useContext(WorkflowApprovalsContext)
  if (!ctx) {
    throw new Error('useWorkflowApprovals must be used within a WorkflowApprovalsProvider')
  }
  return ctx
}
