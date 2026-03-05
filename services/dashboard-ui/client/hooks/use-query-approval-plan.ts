import { useQuery } from '@tanstack/react-query'
import type { TWorkflowStep } from '@/types'
import { useOrg } from './use-org'

export interface IUseQueryApprovalPlan {
  step: TWorkflowStep
}

export function useQueryApprovalPlan({ step }: IUseQueryApprovalPlan) {
  const { org } = useOrg()

  const { data: plan, isLoading, error } = useQuery({
    queryKey: ['approval-plan', org?.id, step?.id, step?.approval?.id],
    queryFn: async () => {
      const res = await fetch(
        `/api/orgs/${org!.id}/workflows/${step.install_workflow_id}/steps/${step.id}/approvals/${step.approval!.id}/contents`
      )
      if (!res.ok) {
        throw new Error(`Failed to fetch approval contents: ${res.status}`)
      }
      return res.json()
    },
    enabled: !!org?.id && !!step?.id && !!step?.install_workflow_id && !!step?.approval?.id,
  })

  return { plan, isLoading, error: error?.error }
}
