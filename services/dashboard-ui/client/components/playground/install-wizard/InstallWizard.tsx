import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { MainLayout } from '@/components/layout/MainLayout'
import { MainTopbar } from '@/components/layout/MainTopbar'
import { OnboardingWizardLayout } from '@/components/onboarding/OnboardingWizard/OnboardingWizard'
import { BreadcrumbNav } from '@/components/navigation/Breadcrumb'
import { WorkflowSteps } from '@/components/workflows/WorkflowSteps/WorkflowSteps'
import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'
import {
  OnboardingWizardProvider,
  type IWizardStepComponentProps,
  type IWizardStepDef,
} from '@/providers/onboarding-wizard-provider'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import { WorkflowContext } from '@/providers/workflow-provider'
import type { TNavLink, TNuonVersion, TWorkflow, TWorkflowStep } from '@/types'
import { cn } from '@/utils/classnames'

type TMode = 'create' | 'teardown'

const INSTALL_NAME = 'acme-payments'

const versions: TNuonVersion = {
  api: { git_ref: 'abc1234', version: '1.2.3' },
  ui: { version: '4.5.6' },
}

const mockNotifications = {
  emitNotification: async () => false,
  permission: 'default' as NotificationPermission,
  requestPermission: async () => 'default' as NotificationPermission,
  isSupported: false,
  settings: { permissionRequested: false },
  hasRequestedPermission: false,
  muted: false,
  toggleMute: () => {},
}

const createCrumbs: TNavLink[] = [
  { path: '/org-1', text: 'Acme' },
  { path: '/org-1/apps', text: 'Apps' },
  { path: '/org-1/apps/app-1', text: 'payments' },
  { path: '/org-1/apps/app-1/installs/new', text: 'Create install' },
]

const teardownCrumbs: TNavLink[] = [
  { path: '/org-1', text: 'Acme' },
  { path: '/org-1/apps', text: 'Apps' },
  { path: '/org-1/apps/app-1', text: 'payments' },
  { path: '/org-1/apps/app-1/installs/inst-1', text: INSTALL_NAME },
  { path: '/org-1/apps/app-1/installs/inst-1/teardown', text: 'Tear down' },
]

const PROVISION_STEPS = [
  'generate install state',
  'generate install stack',
  'await install stack',
  'runner healthy',
  'provision sandbox plan',
  'provision sandbox apply',
  'deploy components',
]

const TEARDOWN_STEPS = [
  'teardown components',
  'deprovision sandbox plan',
  'deprovision sandbox apply',
  'deprovision runner',
]

interface IWizardActions {
  stepMs: number
  exitToInstall: () => void
  forgetInstall: () => void
}

const WizardActionsContext = createContext<IWizardActions>({
  stepMs: 1200,
  exitToInstall: () => {},
  forgetInstall: () => {},
})

const readName = (sharedData: Record<string, unknown>) =>
  typeof sharedData.name === 'string' ? sharedData.name.trim() : ''

const toWorkflowStep = (
  name: string,
  index: number,
  progress: number,
  workflowId: string
): TWorkflowStep => {
  const status = index < progress ? 'success' : index === progress ? 'in-progress' : 'pending'
  return {
    id: `${workflowId}-step-${index}`,
    name,
    execution_type: 'system',
    step_target_type: 'install_deploys',
    step_target_id: `${workflowId}-target-${index}`,
    install_workflow_id: workflowId,
    owner_id: 'inst-1',
    group_idx: index,
    finished: status === 'success',
    started_at: status === 'pending' ? undefined : '2026-01-01T00:00:00Z',
    execution_time: status === 'success' ? (index + 2) * 4_000_000_000 : 0,
    status: { status, history: [] },
  } as TWorkflowStep
}

const useSimulatedWorkflow = (names: string[], stepMs: number) => {
  const [progress, setProgress] = useState(0)
  const finished = progress >= names.length

  useEffect(() => {
    if (finished) return
    const timer = setTimeout(() => setProgress((current) => current + 1), stepMs)
    return () => clearTimeout(timer)
  }, [progress, finished, stepMs])

  return { progress, finished }
}

const GoToInstallButton = () => {
  const { exitToInstall } = useContext(WizardActionsContext)

  return (
    <Button type="button" variant="secondary" onClick={exitToInstall}>
      Go to install
    </Button>
  )
}

const StepActions = ({
  onGoBack,
  onAdvance,
  label,
  disabled,
  disabledReason,
  showCaret = false,
  variant = 'primary',
  secondary,
}: {
  onGoBack?: () => void
  onAdvance: () => void
  label: string
  disabled?: boolean
  disabledReason?: string
  showCaret?: boolean
  variant?: 'primary' | 'danger'
  secondary?: ReactNode
}) => (
  <div className={cn('flex flex-wrap gap-3', onGoBack ? 'justify-between' : 'justify-end')}>
    {onGoBack ? (
      <Button type="button" variant="secondary" onClick={onGoBack}>
        <Icon variant="CaretLeftIcon" weight="bold" /> Back
      </Button>
    ) : null}
    <div className="flex flex-wrap items-center gap-3">
      {secondary}
      <Button
        type="button"
        variant={variant}
        disabled={disabled}
        onClick={onAdvance}
        tooltipProps={disabled && disabledReason ? { tipContent: disabledReason } : undefined}
      >
        {label}
        {showCaret ? <Icon variant="CaretRightIcon" weight="bold" /> : null}
      </Button>
    </div>
  </div>
)

const ChooseAppStep = ({ onAdvance }: IWizardStepComponentProps) => (
  <div className="flex flex-col gap-8">
    <Card elevation="1">
      <Text theme="neutral">App picker goes here.</Text>
      <div className="flex items-center justify-between gap-4">
        <Code variant="inline">payments</Code>
        <Text variant="subtext" theme="neutral">Selected</Text>
      </div>
    </Card>
    <StepActions onAdvance={onAdvance} label="Continue" showCaret />
  </div>
)

const ChooseCloudStep = ({ onAdvance, onGoBack }: IWizardStepComponentProps) => (
  <div className="flex flex-col gap-8">
    <Card elevation="1">
      <Text theme="neutral">Cloud account picker goes here.</Text>
      <div className="flex items-center justify-between gap-4">
        <Code variant="inline">acme-prod</Code>
        <Text variant="subtext">AWS</Text>
      </div>
    </Card>
    <StepActions onGoBack={onGoBack} onAdvance={onAdvance} label="Continue" showCaret />
  </div>
)

const ConfigureStep = ({
  onAdvance,
  onGoBack,
  setSharedData,
  sharedData,
}: IWizardStepComponentProps) => {
  const name = typeof sharedData.name === 'string' ? sharedData.name : ''

  return (
    <div className="flex flex-col gap-8">
      <Input
        id="install-name"
        labelProps={{ labelText: 'Install name' }}
        value={name}
        autoComplete="off"
        onChange={(e) => setSharedData('name', e.target.value)}
      />
      <StepActions
        onGoBack={onGoBack}
        onAdvance={onAdvance}
        label="Continue"
        showCaret
        disabled={!name.trim()}
        disabledReason="Enter an install name"
      />
    </div>
  )
}

const ReviewCreateStep = ({ onAdvance, onGoBack, sharedData }: IWizardStepComponentProps) => (
  <div className="flex flex-col gap-8">
    <Card elevation="1">
      <div className="grid gap-4 sm:grid-cols-2">
        <LabeledValue label="App">
          <Code variant="inline">payments</Code>
        </LabeledValue>
        <LabeledValue label="Cloud account">
          <Code variant="inline">acme-prod</Code>
        </LabeledValue>
        <LabeledValue label="Install">
          <Code variant="inline">{readName(sharedData) || 'unnamed'}</Code>
        </LabeledValue>
      </div>
    </Card>
    <StepActions onGoBack={onGoBack} onAdvance={onAdvance} label="Create install" />
  </div>
)

const ReviewTeardownStep = ({ onAdvance }: IWizardStepComponentProps) => (
  <div className="flex flex-col gap-8">
    <Card elevation="1">
      <Text theme="neutral">Teardown summary goes here.</Text>
      <LabeledValue label="Install">
        <Code variant="inline">{INSTALL_NAME}</Code>
      </LabeledValue>
    </Card>
    <StepActions onAdvance={onAdvance} label="Continue" showCaret secondary={<GoToInstallButton />} />
  </div>
)

const ConfirmTeardownStep = ({ onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const [name, setName] = useState('')
  const matches = name === INSTALL_NAME

  return (
    <div className="flex flex-col gap-8">
      <Text>
        Type <Code variant="inline">{INSTALL_NAME}</Code> to confirm teardown.
      </Text>
      <Input
        id="confirm-install-name"
        labelProps={{ labelText: 'Install name' }}
        value={name}
        autoComplete="off"
        error={name.length > 0 && !matches}
        errorMessage={name.length > 0 && !matches ? "Install name doesn't match" : undefined}
        onChange={(e) => setName(e.target.value)}
      />
      <StepActions
        onGoBack={onGoBack}
        onAdvance={onAdvance}
        label="Start teardown"
        variant="danger"
        disabled={!matches}
        disabledReason="Install name doesn't match"
        secondary={<GoToInstallButton />}
      />
    </div>
  )
}

const WorkflowStep = ({ mode, stepId }: { mode: TMode; stepId: string }) => {
  const { stepMs, exitToInstall, forgetInstall } = useContext(WizardActionsContext)
  const { markComplete, sharedData } = useOnboardingWizard()
  const names = mode === 'create' ? PROVISION_STEPS : TEARDOWN_STEPS
  const workflowId = mode === 'create' ? 'wf-provision' : 'wf-deprovision'
  const { progress, finished } = useSimulatedWorkflow(names, stepMs)
  const installName = mode === 'create' ? readName(sharedData) || 'acme-west' : INSTALL_NAME

  useEffect(() => {
    if (finished) markComplete(stepId)
  }, [finished, markComplete, stepId])

  const workflowSteps = names.map((name, index) =>
    toWorkflowStep(name, index, progress, workflowId)
  )
  const workflow = {
    id: workflowId,
    owner_id: 'inst-1',
    type: mode === 'create' ? 'provision' : 'deprovision',
    status: { status: finished ? 'success' : 'in-progress' },
  } as TWorkflow

  return (
    <div className="flex flex-col gap-6">
      {finished ? (
        <Banner theme="success">
          {mode === 'create' ? (
            <Text>
              <Code variant="inline">{installName}</Code> is provisioned.
            </Text>
          ) : (
            <Text>
              <Code variant="inline">{installName}</Code> is removed from the cloud account. Forget
              the install to remove it from Nuon.
            </Text>
          )}
        </Banner>
      ) : (
        <Text theme="neutral" flex className="gap-2">
          <Icon variant="Loading" />
          {mode === 'create' ? 'Provisioning' : 'Tearing down'}
          <Code variant="inline">{installName}</Code>
        </Text>
      )}

      <WorkflowContext.Provider
        value={{
          workflow,
          stopPolling: () => {},
          workflowSteps,
          hasApprovals: false,
          failedSteps: [],
          pendingApprovals: [],
          discardedSteps: [],
          completedSteps: workflowSteps.filter((step) => step.finished),
          stepsWithPolicyViolations: [],
          totalSteps: workflowSteps.length,
          pendingApprovalsCount: 0,
          discardedStepsCount: 0,
          completedStepsCount: Math.min(progress, names.length),
          failedStepsCount: 0,
          policyViolationsCount: 0,
        }}
      >
        <WorkflowSteps workflowSteps={workflowSteps} />
      </WorkflowContext.Provider>

      <div className="flex flex-wrap justify-end gap-3">
        {mode === 'teardown' || !finished ? <GoToInstallButton /> : null}
        {finished ? (
          mode === 'create' ? (
            <Button variant="primary" onClick={exitToInstall}>
              View install <Icon variant="CaretRightIcon" weight="bold" />
            </Button>
          ) : (
            <Button variant="danger" onClick={forgetInstall}>
              <Icon variant="TrashIcon" /> Forget install
            </Button>
          )
        ) : null}
      </div>
    </div>
  )
}

const ProvisionStep = () => <WorkflowStep mode="create" stepId="provision" />
const TeardownStep = () => <WorkflowStep mode="teardown" stepId="teardown" />

const createSteps: IWizardStepDef[] = [
  {
    id: 'app',
    navLabel: 'App',
    title: 'Choose an app',
    description: 'Choose the app template this install runs.',
    component: ChooseAppStep,
  },
  {
    id: 'cloud',
    navLabel: 'Cloud',
    title: 'Cloud account',
    description: 'Choose the cloud account this install deploys into.',
    component: ChooseCloudStep,
  },
  {
    id: 'configure',
    navLabel: 'Configure',
    title: 'Configure',
    description: 'Name the install and set its inputs.',
    component: ConfigureStep,
  },
  {
    id: 'review',
    navLabel: 'Review',
    title: 'Review',
    description: 'Confirm the install before creating it.',
    component: ReviewCreateStep,
  },
  {
    id: 'provision',
    navLabel: 'Provision',
    title: 'Provision',
    description: 'The provision workflow builds the runner, sandbox, and components in the cloud account.',
    lockPrevious: true,
    component: ProvisionStep,
  },
]

const teardownSteps: IWizardStepDef[] = [
  {
    id: 'review',
    navLabel: 'Review',
    title: 'Review teardown',
    description: 'Teardown removes this install from the cloud account.',
    component: ReviewTeardownStep,
  },
  {
    id: 'confirm',
    navLabel: 'Confirm',
    title: 'Confirm',
    description: 'Confirm teardown before it starts.',
    component: ConfirmTeardownStep,
  },
  {
    id: 'teardown',
    navLabel: 'Teardown',
    title: 'Teardown',
    description: 'The teardown workflow removes the components, sandbox, and runner from the cloud account.',
    lockPrevious: true,
    component: TeardownStep,
  },
]

const TopbarAction = ({ mode, onCancel }: { mode: TMode; onCancel: () => void }) => {
  const { steps, currentStepIndex } = useOnboardingWizard()
  if (mode === 'teardown' || steps[currentStepIndex]?.lockPrevious) return null

  return (
    <Button variant="ghost" size="sm" className="shrink-0" onClick={onCancel}>
      Cancel
    </Button>
  )
}

const PaneMessage = ({
  title,
  description,
  actionLabel,
  onAction,
}: {
  title: string
  description: string
  actionLabel: string
  onAction: () => void
}) => (
  <div className="h-full overflow-y-auto px-6 pt-14 pb-8">
    <div className="max-w-4xl mx-auto flex flex-col gap-6">
      <Text variant="h2" role="heading" level={2}>
        {title}
      </Text>
      <Text theme="neutral">{description}</Text>
      <div>
        <Button variant="secondary" onClick={onAction}>
          {actionLabel}
        </Button>
      </div>
    </div>
  </div>
)

type TPhase = 'wizard' | 'cancelled' | 'install' | 'forgotten'

const RESULTS: Record<Exclude<TPhase, 'wizard'>, { title: string; description: string }> = {
  cancelled: {
    title: 'Create install canceled',
    description: 'Nothing was changed. The real wizard returns to the page it was opened from.',
  },
  install: {
    title: 'Install overview',
    description: 'The real wizard opens the install overview here.',
  },
  forgotten: {
    title: 'Install forgotten',
    description: 'The real wizard returns to the installs list here.',
  },
}

export interface IInstallWizard {
  mode: TMode
  startAtWorkflow?: boolean
  stepMs?: number
}

export const InstallWizard = ({ mode, startAtWorkflow = false, stepMs = 1200 }: IInstallWizard) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(true)
  const [phase, setPhase] = useState<TPhase>('wizard')
  const [session, setSession] = useState(0)
  const steps = mode === 'create' ? createSteps : teardownSteps

  const restart = () => {
    setSession((current) => current + 1)
    setPhase('wizard')
  }

  const actions: IWizardActions = {
    stepMs,
    exitToInstall: () => setPhase('install'),
    forgetInstall: () => setPhase('forgotten'),
  }

  const result = phase === 'wizard' ? null : RESULTS[phase]

  return (
    <NotificationContext.Provider value={mockNotifications}>
      <BreadcrumbContext.Provider
        value={{
          breadcrumbLinks: mode === 'create' ? createCrumbs : teardownCrumbs,
          isLoading: false,
          updateBreadcrumb: () => {},
        }}
      >
        <SidebarContext.Provider
          value={{
            isSidebarOpen,
            closeSidebar: () => setIsSidebarOpen(false),
            openSidebar: () => setIsSidebarOpen(true),
            toggleSidebar: () => setIsSidebarOpen((open) => !open),
          }}
        >
          <WizardActionsContext.Provider value={actions}>
            <OnboardingWizardProvider
              key={session}
              steps={steps}
              initialStepIndex={startAtWorkflow ? steps.length - 1 : 0}
              initialSharedData={startAtWorkflow && mode === 'create' ? { name: 'acme-west' } : undefined}
              onComplete={actions.exitToInstall}
            >
              <div className="flex h-full min-h-0">
                <MainLayout versions={versions}>
                  <main className="flex flex-col flex-1 min-h-0 w-full">
                    <MainTopbar>
                      <BreadcrumbNav />
                      {phase === 'wizard' ? (
                        <TopbarAction mode={mode} onCancel={() => setPhase('cancelled')} />
                      ) : null}
                    </MainTopbar>
                    <div className="relative flex-1 min-h-0">
                      <div className="absolute inset-0">
                        {result ? (
                          <PaneMessage
                            title={result.title}
                            description={result.description}
                            actionLabel="Start over"
                            onAction={restart}
                          />
                        ) : (
                          <OnboardingWizardLayout variant="pane" />
                        )}
                      </div>
                    </div>
                  </main>
                </MainLayout>
              </div>
            </OnboardingWizardProvider>
          </WizardActionsContext.Provider>
        </SidebarContext.Provider>
      </BreadcrumbContext.Provider>
    </NotificationContext.Provider>
  )
}
