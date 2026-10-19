import { useLocation, useParams, useSearchParams } from 'react-router'
import { useContext, useEffect, useRef, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import type { IPanel } from '@/components/surfaces/Panel'
import { useInstallLink, useInstallNested } from '@/hooks/use-install-path'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useWorkflow } from '@/hooks/use-workflow'
import {
  isWorkflowPath,
  WorkflowInstallContext,
  workflowPageSuffix,
} from '@/lib/workflow-href'
import { OrgContext } from '@/providers/org-provider'
import { getWorkflowStep } from '@/lib'
import type { TWorkflowStep } from '@/types'
import { ActionRunStepDetails } from '../action-run-details/ActionRunStepDetails'
import { DeployStepDetails } from '../deploy-details/DeployStepDetails'
import { InstallGroupStepDetails } from '../install-group-details'
import { SandboxRunStepDetails } from '../sandbox-run-details/SandboxRunStepDetails'
import { StackStepDetails } from '../stack-details/StackStepDetails'
import { RunnerStepDetails } from '../RunnerStepDetails'
import { VerifyHealthStepDetails } from '../verify-health-details'
import { SyncSecretsStepDetails } from '../SyncSecretsStepDetails'
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

export interface IStepDetailPanelContainer extends IPanel {
  children: ReactNode
  initStep: TWorkflowStep
  planOnly?: boolean
  pollInterval?: number
  shouldPoll?: boolean
}

export const StepDetailPanelContainer = ({
  children,
  initStep,
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
  )
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
  const { addPanel, addPanelIfEmpty, removePanel } = useSurfaces()
  const { workflow } = useWorkflow()
  const org = useContext(OrgContext)?.org
  const params = useParams()
  const { pathname } = useLocation()
  const installLink = useInstallLink()
  const nested = useInstallNested()
  const scopedInstallId = useContext(WorkflowInstallContext)
  const [searchParams] = useSearchParams()
  const workflowId = step.install_workflow_id || workflow?.id
  const installId =
    scopedInstallId ||
    params.installId ||
    workflow?.metadata?.install_id ||
    workflow?.owner_id
  const href =
    org?.id && installId && workflowId && step.id
      ? installLink({
          orgId: org.id,
          appId: params.appId,
          installId,
          suffix: workflowPageSuffix(workflowId, nested, step.id),
        })
      : ''
  const onWorkflowPage = isWorkflowPath(pathname, workflowId)
  const autoOpened = useRef(false)
  const autoOpenedPanelIdRef = useRef<string | null>(null)

  const panel = (
    <StepDetailPanelContainer
      panelKey={step.id}
      initStep={step}
      size={getStepPanelSize(step)}
      shouldPoll
      planOnly={planOnly}
    >
      {getStepPanelDetails(step)}
    </StepDetailPanelContainer>
  )

  const handleAddPanel = () => {
    addPanel(panel, step.id)
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
      step.id === searchParams?.get('panel')
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
      autoOpenedPanelIdRef.current = addPanelIfEmpty(panel, step.id) ?? null
    }
  }, [
    workflow?.id,
    workflowBlocked,
    suppressAutoOpen,
    isPendingApproval,
    isPendingAwaitStack,
  ])

  useEffect(() => {
    if (autoOpenedPanelIdRef.current && step.finished) {
      removePanel(autoOpenedPanelIdRef.current, step.id)
      autoOpenedPanelIdRef.current = null
    }
  }, [step.finished])

  if (href) {
    return (
      <Link
        href={href}
        onClick={(event) => {
          if (!onWorkflowPage) return
          if (
            event.metaKey ||
            event.ctrlKey ||
            event.shiftKey ||
            event.altKey ||
            event.button !== 0
          ) {
            return
          }
          event.preventDefault()
          handleAddPanel()
        }}
      >
        View details
      </Link>
    )
  }

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
