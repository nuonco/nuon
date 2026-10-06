import { useCallback, useEffect, useRef } from 'react'
import { useParams, useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IPanel } from '@/components/surfaces/Panel'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { useSearchParamState } from '@/hooks/use-search-param-state'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getBranchWorkflowRun } from '@/lib'
import { isActiveStepStatus } from '@/components/branches/shared/step-status'
import { getRunTitle } from '@/components/branches/shared/run-title'
import { scrollElementIntoView } from '@/utils/scroll'
import type { TInstallWorkflowStep } from '@/types'
import { WorkflowRunPanel } from './WorkflowRunPanel'

interface IWorkflowRunPanelContainer extends IPanel {
  onClose: () => void
  runId?: string
}

const INTERNAL_STEP_NAMES = new Set([
  'check ignored changes',
  'setup preview',
  'preview install impact',
])

export const filterWorkflowPanelSteps = (steps: TInstallWorkflowStep[]) =>
  steps.filter(
    (step) =>
      step.owner_type !== 'components' &&
      step.execution_type !== 'hidden' &&
      !INTERNAL_STEP_NAMES.has(step.name)
  )

export const WorkflowRunPanelContainer = ({
  onClose,
  runId: runIdProp,
  ...props
}: IWorkflowRunPanelContainer) => {
  const params = useParams()
  const { org } = useOrg()
  const orgId = org?.id ?? (params.orgId as string)
  const appId = params.appId as string
  const branchId = params.branchId as string
  const runId = runIdProp ?? (params.runId as string)

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

  const steps = filterWorkflowPanelSteps(run?.steps || [])
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

const useClearWorkflowParams = () => {
  const [, setSearchParams] = useSearchParams()
  return useCallback(() => {
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
}

export const useOpenWorkflowRunPanel = () => {
  const [, setSearchParams] = useSearchParams()
  return useCallback(
    (runId: string, stepId?: string) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev)
          next.set('workflow', runId)
          if (stepId) next.set('step', stepId)
          else next.delete('step')
          return next
        },
        { replace: true }
      )
    },
    [setSearchParams]
  )
}

const useWorkflowRunPanelSync = (
  runIdFilter: string | null,
  enabled: boolean
) => {
  const { addPanel, removePanel, panels } = useSurfaces()
  const [searchParams] = useSearchParams()
  const workflowParam = searchParams.get('workflow')
  const clearParams = useClearWorkflowParams()
  const panelRef = useRef<{ id: string; runId: string } | null>(null)
  const openPanel = panels.some((p) => p?.id === panelRef.current?.id)
    ? panelRef.current
    : null

  useEffect(() => {
    const target =
      enabled && workflowParam && (!runIdFilter || workflowParam === runIdFilter)
        ? workflowParam
        : null

    if (openPanel && openPanel.runId !== target) {
      removePanel(openPanel.id)
      panelRef.current = null
    }
    if (target && openPanel?.runId !== target) {
      panelRef.current = {
        runId: target,
        id: addPanel(
          <WorkflowRunPanelContainer runId={target} onClose={clearParams} />
        ),
      }
    }
  }, [
    enabled,
    workflowParam,
    runIdFilter,
    openPanel,
    addPanel,
    removePanel,
    clearParams,
  ])
}

export const WorkflowRunPanelHost = () => {
  useWorkflowRunPanelSync(null, true)
  return null
}

export const WorkflowRunPanelButton = ({ runId }: { runId: string }) => {
  const hasNewAppIA = useNewAppIA()
  const openWorkflowRunPanel = useOpenWorkflowRunPanel()
  useWorkflowRunPanelSync(runId, !hasNewAppIA)

  const openPanel = () => openWorkflowRunPanel(runId)

  return (
    <Button variant="secondary" onClick={openPanel}>
      <Icon variant="ListChecksIcon" size={16} />
      Workflow steps
    </Button>
  )
}
