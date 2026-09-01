import { useCallback, useEffect, useRef } from 'react'
import { useParams, useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IPanel } from '@/components/surfaces/Panel'
import { useOrg } from '@/hooks/use-org'
import { useSearchParamState } from '@/hooks/use-search-param-state'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getBranchWorkflowRun } from '@/lib'
import { isActiveStepStatus } from '@/components/branches/shared/step-status'
import { getRunTitle } from '@/components/branches/shared/run-title'
import { scrollElementIntoView } from '@/utils/scroll'
import { WorkflowRunPanel } from './WorkflowRunPanel'

interface IWorkflowRunPanelContainer extends IPanel {
  onClose: () => void
}

export const WorkflowRunPanelContainer = ({
  onClose,
  ...props
}: IWorkflowRunPanelContainer) => {
  const params = useParams()
  const { org } = useOrg()
  const orgId = org?.id ?? (params.orgId as string)
  const appId = params.appId as string
  const branchId = params.branchId as string
  const runId = params.runId as string

  const [urlStepId, setUrlStepId] = useSearchParamState('step')
  const stepDetailRef = useRef<HTMLDivElement>(null)
  const pendingScrollRef = useRef(false)

  const { data: run, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-run', orgId, appId, branchId, runId],
    queryFn: () => getBranchWorkflowRun({ orgId, appId, branchId, runId }),
    enabled: !!orgId && !!appId && !!branchId && !!runId,
    refetchInterval: 5000,
  })

  const steps = (run?.steps || []).filter((s) => s.owner_type !== 'components')
  const activeStep = steps.find((step) =>
    isActiveStepStatus(step.status?.status)
  )
  const urlStep = urlStepId
    ? (steps.find((s) => s.id === urlStepId) ?? null)
    : null
  const selectedStep = urlStep ?? activeStep ?? steps[0] ?? null
  const selectedStepId = selectedStep?.id ?? null
  const branchRun = run?.app_branch_runs?.at(0)

  useEffect(() => {
    if (pendingScrollRef.current) {
      scrollElementIntoView(stepDetailRef.current, { block: 'start' })
      pendingScrollRef.current = false
    }
  }, [selectedStepId])

  const handleJumpToActive = () => {
    if (!activeStep) return
    if (selectedStepId === activeStep.id) {
      scrollElementIntoView(stepDetailRef.current, { block: 'start' })
      return
    }
    pendingScrollRef.current = true
    setUrlStepId(activeStep.id ?? null)
  }

  return (
    <WorkflowRunPanel
      {...props}
      onClose={onClose}
      isLoading={isLoading}
      steps={steps}
      selectedStep={selectedStep}
      activeStep={activeStep}
      onSelectStep={(step) => setUrlStepId(step?.id ?? null)}
      onJumpToActive={handleJumpToActive}
      appBranchId={branchId}
      appBranchRunId={branchRun?.id}
      stepDetailRef={stepDetailRef}
      runTitle={getRunTitle(run)}
      status={run?.status?.status || 'unknown'}
    />
  )
}

export const WorkflowRunPanelButton = ({ runId }: { runId: string }) => {
  const { addPanel, removePanel, panels } = useSurfaces()
  const [searchParams, setSearchParams] = useSearchParams()
  const workflowParam = searchParams.get('workflow')
  const panelIdRef = useRef<string | null>(null)
  const openPanelId =
    panels.find((p) => p?.id === panelIdRef.current)?.id ?? null

  const clearParams = useCallback(() => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        next.delete('workflow')
        next.delete('step')
        return next
      },
      { replace: true }
    )
  }, [setSearchParams])

  useEffect(() => {
    const shouldOpen = workflowParam === runId
    if (shouldOpen && !openPanelId) {
      panelIdRef.current = addPanel(
        <WorkflowRunPanelContainer onClose={clearParams} />
      )
    } else if (!shouldOpen && openPanelId) {
      removePanel(openPanelId)
      panelIdRef.current = null
    }
  }, [workflowParam, runId, openPanelId, addPanel, removePanel, clearParams])

  const openPanel = () => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        next.set('workflow', runId)
        return next
      },
      { replace: true }
    )
  }

  return (
    <Button variant="secondary" onClick={openPanel}>
      <Icon variant="ListChecksIcon" size={16} />
      Workflow steps
    </Button>
  )
}
