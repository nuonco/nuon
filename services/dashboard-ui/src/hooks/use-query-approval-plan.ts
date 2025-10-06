'use client'

import { useEffect, useState } from 'react'
import type { TWorkflowStep } from '@/types'
import { useOrg } from './use-org'

export interface IUseQueryApprovalPlan {
  step: TWorkflowStep
}

export function useQueryApprovalPlan({ step }: IUseQueryApprovalPlan) {
  const { org } = useOrg()
  const [plan, setPlan] = useState<any>()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

  useEffect(() => {
    if (
      !org?.id ||
      !step?.id ||
      !step?.install_workflow_id ||
      !step?.approval?.id
    ) {
      setIsLoading(false)
      setError('Missing required org or step properties.')
      return
    }

    fetch(
      `/api/orgs/${org.id}/workflows/${step.workflow_id}/steps/${step.id}/approvals/${step.approval.id}/contents`
    )
      .then((r) => r.json())
      .then((res) => {
        setPlan(res)
        setIsLoading(false)
      })
      .catch((err) => {
        setIsLoading(false)
        setError(err?.message || 'Unable to get approval plan')
      })
  }, [org?.id, step?.id, step?.workflow_id, step?.approval?.id])

  return { plan, isLoading, error }
}
