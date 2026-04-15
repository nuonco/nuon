import { useMemo } from 'react'
import type { TWorkflow } from '@/types'

export const useWorkflowActions = (workflow: TWorkflow, hasApprovals: boolean) => {
  return useMemo(() => {
    const isFinished = workflow?.finished
    const status = workflow?.status?.status
    const isCancelled = status === 'cancelled'
    const isError = status === 'error'
    const isPlanOnly = workflow?.plan_only
    const hasApprovalPrompt = workflow?.approval_option === 'prompt'

    const canShowApproveAll =
      hasApprovalPrompt &&
      !isFinished &&
      !isPlanOnly &&
      !isCancelled &&
      hasApprovals

    const canShowCancel =
      !isFinished &&
      !isCancelled &&
      !isError

    return {
      canShowApproveAll,
      canShowCancel,
    }
  }, [workflow, hasApprovals])
}
