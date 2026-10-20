import { useSearchParams } from 'react-router'
import { useCallback, useContext, useEffect, useRef, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Loading } from '@/components/common/Loading'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useWorkflow } from '@/hooks/use-workflow'
import { getWorkflowStep } from '@/lib'
import { InstallAppConfigContext, InstallAppConfigProvider } from '@/providers/install-app-config-provider'
import { InstallContext, InstallProvider } from '@/providers/install-provider'
import { hasPanelKey } from '@/providers/surfaces-provider'
import type { TWorkflowStep } from '@/types'
import { ActionRunStepDetails } from '../action-run-details/ActionRunStepDetails'
import { DeployStepDetails } from '../deploy-details/DeployStepDetails'
import { InstallGroupStepDetails } from '../install-group-details'
import { SandboxRunStepDetails } from '../sandbox-run-details/SandboxRunStepDetails'
import { StackStepDetails } from '../stack-details/StackStepDetails'
import { RunnerStepDetails } from '../RunnerStepDetails'
import { VerifyHealthStepDetails } from '../verify-health-details'
import { SyncSecretsStepDetails } from '../SyncSecretsStepDetails'
import { StepTitle } from '../StepTitle'
import { StepDetailPanel } from './StepDetailPanel'

type TPanelSize = IPanel['size']

export function getStepPanelSize(step: TWorkflowStep): TPanelSize {
  if (
    step?.step_target_type === 'install_deploys' ||
    step?.step_target_type === 'install_sandbox_runs' ||
    step?.step_target_type === 'install_action_workflow_runs' ||
    step?.step_target_type === 'install_workflow_steps' ||
    step?.step_target_type === 'app_branches'
  ) {
    return '3/4'
  }
  return 'half'
}

export function getStepPanelDetails(step: TWorkflowStep): ReactNode {
  if (step.step_target_type === 'install_action_workflow_runs')
    return <ActionRunStepDetails />
  if (step.step_target_type === 'install_deploys') return <DeployStepDetails />
  if (step.step_target_type === 'install_sandbox_runs')
    return <SandboxRunStepDetails />
  if (step.step_target_type === 'install_stack_versions')
    return <StackStepDetails />
  if (step.step_target_type === 'runners') return <RunnerStepDetails />
  if (step.step_target_type === 'install_components')
    return <VerifyHealthStepDetails />
  if (step.step_target_type === 'install_workflow_steps')
    return <SyncSecretsStepDetails />
  if (step.step_target_type === 'app_branches')
    return <InstallGroupStepDetails />
  return (
    <EmptyState
      variant="history"
      emptyTitle="Waiting for step to start"
      emptyMessage="Details will appear here as soon as the runner picks up this step."
    />
  )
}

// Panels render at the nearest SurfacesProvider, which can sit above the
// InstallProvider the step was opened under (app branch pages, the install
// wizard).
const StepInstallScope = ({
  installId,
  loadingElement,
  children,
}: {
  installId?: string
  loadingElement: ReactNode
  children: ReactNode
}) => {
  const install = useContext(InstallContext)
  const appConfig = useContext(InstallAppConfigContext)
  if (!installId) return <>{children}</>
  if (install?.install?.id === installId) {
    return appConfig ? (
      <>{children}</>
    ) : (
      <InstallAppConfigProvider>{children}</InstallAppConfigProvider>
    )
  }
  return (
    <InstallProvider installId={installId} loadingElement={loadingElement}>
      <InstallAppConfigProvider>{children}</InstallAppConfigProvider>
    </InstallProvider>
  )
}

export interface IStepDetailPanelContainer extends IPanel {
  children: ReactNode
  initStep: TWorkflowStep
  installId?: string
  planOnly?: boolean
  pollInterval?: number
  shouldPoll?: boolean
}

export const StepDetailPanelContainer = ({
  children,
  initStep,
  installId,
  planOnly = false,
  pollInterval = 10000,
  shouldPoll = false,
  ...props
}: IStepDetailPanelContainer) => {
  const { org } = useOrg()

  const { data: step = initStep } = useQuery<TWorkflowStep>({
    queryKey: [
      'workflow-step',
      org?.id,
      initStep.install_workflow_id,
      initStep.id,
    ],
    queryFn: () =>
      getWorkflowStep({
        orgId: org!.id,
        workflowId: initStep.install_workflow_id,
        workflowStepId: initStep.id,
      }),
    refetchInterval: (query) => {
      if (!shouldPoll) return false
      if (query.state.data?.finished) return 30_000
      if (!query.state.data?.step_target_type) return 1_500
      return pollInterval
    },
    initialData: initStep,
    enabled: !!org?.id,
  })

  return (
    <StepInstallScope
      installId={installId}
      loadingElement={
        <Panel heading={<StepTitle step={step} />} size="half" {...props}>
          <Loading />
        </Panel>
      }
    >
      <StepDetailPanel
        step={step}
        planOnly={planOnly}
        triggerHref={
          org?.id && step?.links?.event_wait?.trigger_id
            ? `/${org.id}/settings/triggers/${step.links.event_wait.trigger_id}`
            : undefined
        }
        {...props}
      >
        {children}
      </StepDetailPanel>
    </StepInstallScope>
  )
}

// A step's workflow is scoped to one install, but this builds panels from
// anywhere workflow steps are shown (step rows, error banners, a stopped-at
// link), so every caller needs the same install resolution rather than each
// reinventing it and drifting out of sync with StepInstallScope.
export const useStepDetailPanel = () => {
  const { addPanel } = useSurfaces()
  const { workflow } = useWorkflow()
  const install = useContext(InstallContext)?.install
  const installId =
    install?.id || workflow?.metadata?.install_id || workflow?.owner_id

  const buildPanel = useCallback(
    (step: TWorkflowStep, planOnly = false) => (
      <StepDetailPanelContainer
        panelKey={step.id}
        initStep={step}
        installId={installId}
        size={getStepPanelSize(step)}
        shouldPoll
        planOnly={planOnly}
      >
        {getStepPanelDetails(step)}
      </StepDetailPanelContainer>
    ),
    [installId]
  )

  const openPanel = useCallback(
    (step: TWorkflowStep, planOnly = false) =>
      addPanel(buildPanel(step, planOnly), step.id),
    [addPanel, buildPanel]
  )

  return { buildPanel, openPanel }
}

const StepPanelButton = ({
  step,
  approvalPrompt,
  planOnly = false,
}: {
  approvalPrompt?: boolean
  step: TWorkflowStep
  planOnly?: boolean
}) => {
  const { addPanelIfEmpty, removePanel, panels } = useSurfaces()
  const { workflow } = useWorkflow()
  const { buildPanel, openPanel } = useStepDetailPanel()
  const [searchParams] = useSearchParams()
  const autoOpened = useRef(false)
  const autoOpenedPanelIdRef = useRef<string | null>(null)

  const handleAddPanel = () => {
    openPanel(step, planOnly)
  }

  const isPendingApproval =
    approvalPrompt &&
    step.execution_type === 'approval' &&
    !!step.approval &&
    !step.approval.response &&
    step.status?.status !== 'discarded'

  const isPendingAwaitStack =
    step.step_target_type === 'install_stack_versions' &&
    step.name !== 'generate install stack' &&
    !!step.started_at &&
    !step.finished

  const workflowStatus = workflow?.status?.status
  const workflowBlocked =
    workflowStatus === 'cancelled' || workflowStatus === 'error'
  const suppressAutoOpen =
    planOnly ||
    workflow?.type === 'drift_run' ||
    workflow?.type === 'drift_run_reprovision_sandbox'

  useEffect(() => {
    if (
      !suppressAutoOpen &&
      step.id &&
      hasPanelKey(searchParams, step.id) &&
      !panels.some((p) => p.key === step.id)
    ) {
      handleAddPanel()
    }
  }, [])

  useEffect(() => {
    if (autoOpened.current || !workflow) return
    if (
      !workflowBlocked &&
      !suppressAutoOpen &&
      (isPendingApproval || isPendingAwaitStack)
    ) {
      autoOpened.current = true
      autoOpenedPanelIdRef.current =
        addPanelIfEmpty(buildPanel(step, planOnly), step.id) ?? null
    }
  }, [
    workflow?.id,
    workflowBlocked,
    suppressAutoOpen,
    isPendingApproval,
    isPendingAwaitStack,
    buildPanel,
    step,
    planOnly,
    addPanelIfEmpty,
  ])

  useEffect(() => {
    if (autoOpenedPanelIdRef.current && step.finished) {
      removePanel(autoOpenedPanelIdRef.current, step.id)
      autoOpenedPanelIdRef.current = null
    }
  }, [step.finished])

  return (
    <Button
      className="!text-primary-600 dark:!text-primary-500"
      variant="ghost"
      size="sm"
      onClick={handleAddPanel}
    >
      View details
    </Button>
  )
}

export const StepDetailPanelButton = ({
  step,
  ...props
}: {
  approvalPrompt?: boolean
  step: TWorkflowStep
  planOnly?: boolean
}) => <StepPanelButton step={step} {...props} />
