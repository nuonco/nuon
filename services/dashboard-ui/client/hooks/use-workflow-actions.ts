import { useMemo } from 'react'
import { useAuth } from '@/hooks/use-auth'
import type { TWorkflow } from '@/types'

export const useWorkflowActions = (workflow: TWorkflow, hasApprovals: boolean) => {
  const { user, isLoading } = useAuth()
  
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
    
    const canShowTemporalLink = 
      !isLoading && 
      user?.email?.endsWith('@nuon.co')
    
    return {
      canShowApproveAll,
      canShowCancel,
      canShowTemporalLink,
      user,
      isLoading,
    }
  }, [workflow, hasApprovals, user, isLoading])
}