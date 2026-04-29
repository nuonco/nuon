import { createContext, useState, useRef, useEffect, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useWorkflowMetrics } from '@/hooks/use-workflow-metrics'
import { useOrg } from '@/hooks/use-org'
import { getWorkflow } from '@/lib'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TWorkflow } from '@/types'

interface WorkflowContextValue {
  workflow: TWorkflow
  stopPolling: () => void
  workflowSteps: any[]
  hasApprovals: boolean
  failedSteps: any[]
  pendingApprovals: any[]
  discardedSteps: any[]
  completedSteps: any[]
  stepsWithPolicyViolations: any[]
  totalSteps: number
  pendingApprovalsCount: number
  discardedStepsCount: number
  completedStepsCount: number
  failedStepsCount: number
  policyViolationsCount: number
}

export const WorkflowContext = createContext<WorkflowContextValue | undefined>(undefined)

const FAST_POLL_MS = 250
const SLOW_POLL_MS = 4000
const FAST_POLL_DURATION_MS = 60_000

export const WorkflowProvider = ({
  children,
  workflowId,
  pollInterval = SLOW_POLL_MS,
  shouldPoll = false,
}: {
  children: ReactNode
  workflowId: string
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()
  const [pollingEnabled, setPollingEnabled] = useState(shouldPoll)
  const [fastPoll, setFastPoll] = useState(shouldPoll)
  const mountTime = useRef(Date.now())

  useEffect(() => {
    if (!shouldPoll) return
    const remaining = FAST_POLL_DURATION_MS - (Date.now() - mountTime.current)
    if (remaining <= 0) {
      setFastPoll(false)
      return
    }
    const timer = setTimeout(() => setFastPoll(false), remaining)
    return () => clearTimeout(timer)
  }, [shouldPoll])

  const activePollInterval = pollingEnabled
    ? fastPoll
      ? FAST_POLL_MS
      : pollInterval
    : false

  const { data: workflow, isLoading, error } = useQuery({
    queryKey: ['workflow', org.id!, workflowId],
    queryFn: () => getWorkflow({ orgId: org.id!, workflowId }),
    refetchInterval: activePollInterval,
    enabled: !!org.id && !!workflowId,
  })

  const metrics = useWorkflowMetrics(workflow)

  if (error) return <ProviderError error={error} />

  if (isLoading || !workflow) return <ProviderLoading />

  const value: WorkflowContextValue = {
    workflow,
    stopPolling: () => setPollingEnabled(false),
    ...metrics,
  }

  return (
    <WorkflowContext.Provider value={value}>
      {children}
    </WorkflowContext.Provider>
  )
}
