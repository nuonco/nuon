import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Code } from '@/components/common/Code'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { CloudRegion } from '@/components/common/CloudRegion'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Select, type SelectOption } from '@/components/common/form/Select'
import { Toggle } from '@/components/common/form/Toggle'
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
import type {
  TCloudPlatform,
  TNavLink,
  TNuonVersion,
  TWorkflow,
  TWorkflowStep,
} from '@/types'
import { cn } from '@/utils/classnames'
import { regionOptions } from '@/components/onboarding/first-run/constants'

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
  { path: '/org-1/apps/app-1/installs/inst-1/teardown', text: 'Teardown' },
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

const StepModeContext = createContext({ readOnly: false })

type TAppReadiness =
  | 'not-provisionable'
  | 'no-valid-config'
  | 'no-components'
  | 'no-component-builds'

const READINESS_LABEL: Record<TAppReadiness, string> = {
  'not-provisionable': 'Not provisionable',
  'no-valid-config': 'No valid config',
  'no-components': 'No components',
  'no-component-builds': 'No component builds',
}

interface IAppOption {
  id: string
  name: string
  platform: TCloudPlatform
  source: string
  updatedLabel: string
  readiness?: TAppReadiness
}

const APPS: IAppOption[] = [
  {
    id: 'payments',
    name: 'Payments API',
    platform: 'aws',
    source: 'acme/payments',
    updatedLabel: 'synced 8 minutes ago',
  },
  {
    id: 'support-portal',
    name: 'Support portal',
    platform: 'aws',
    source: 'acme/support-portal',
    updatedLabel: 'synced yesterday',
  },
  {
    id: 'edge-gateway',
    name: 'Edge gateway',
    platform: 'gcp',
    source: 'acme/edge-gateway',
    updatedLabel: 'synced 1 hour ago',
    readiness: 'no-components',
  },
  {
    id: 'analytics',
    name: 'Analytics warehouse',
    platform: 'azure',
    source: 'acme/analytics',
    updatedLabel: 'synced 3 days ago',
    readiness: 'no-valid-config',
  },
]

const AppOptionName = ({ app }: { app: IAppOption }) => (
  <span className="flex min-w-0 items-center gap-2">
    <span className="truncate">{app.name}</span>
    {app.readiness ? (
      <Badge size="sm" className="shrink-0">
        {READINESS_LABEL[app.readiness]}
      </Badge>
    ) : null}
  </span>
)

const AppOptionDetail = ({ app }: { app: IAppOption }) => (
  <span className="flex min-w-0 items-center gap-1.5">
    <CloudPlatform
      platform={app.platform}
      displayVariant="icon-only"
      iconSize="14"
      variant="subtext"
      className="shrink-0"
    />
    <span aria-hidden>·</span>
    <Icon variant="GithubLogoIcon" size={12} className="shrink-0" />
    <span className="truncate">{app.source}</span>
    <span aria-hidden>·</span>
    <span className="shrink-0">{app.updatedLabel}</span>
  </span>
)

const APP_OPTIONS: SelectOption[] = APPS.map((app) => ({
  value: app.id,
  label: app.name,
  content: <AppOptionName app={app} />,
  detail: <AppOptionDetail app={app} />,
  searchText: `${app.id} ${app.source}`,
  disabled: app.readiness === 'not-provisionable' || app.readiness === 'no-valid-config',
}))

type TGroupKind = 'labels' | 'wildcard' | 'default'

interface IInstallGroup {
  id: string
  name: string
  kind: TGroupKind
  labels: Record<string, string>
}

interface IAppBranch {
  id: string
  name: string
  groups: IInstallGroup[]
}

const BRANCHES: IAppBranch[] = [
  {
    id: 'main',
    name: 'main',
    groups: [
      {
        id: 'production',
        name: 'Production',
        kind: 'labels',
        labels: { env: 'production', tier: 'critical' },
      },
      {
        id: 'preview',
        name: 'Preview',
        kind: 'wildcard',
        labels: { env: 'preview', pull_request: '*' },
      },
    ],
  },
  {
    id: 'release',
    name: 'release',
    groups: [
      {
        id: 'stable',
        name: 'Stable',
        kind: 'labels',
        labels: { env: 'stable' },
      },
      {
        id: 'remaining',
        name: 'Remaining',
        kind: 'default',
        labels: {},
      },
    ],
  },
]

const appLabel = (value: unknown) => APPS.find((app) => app.id === value)?.name

const groupName = (branchId: unknown, groupId: unknown) =>
  BRANCHES.find((branch) => branch.id === branchId)?.groups.find((group) => group.id === groupId)
    ?.name

const groupDescription = (group: IInstallGroup) => {
  if (group.kind === 'default') return 'This group includes all remaining installs automatically.'
  if (group.kind === 'wildcard') return 'Wildcard selectors cannot be joined by applying labels.'
  const labels = Object.entries(group.labels)
    .map(([key, value]) => `${key}=${value}`)
    .join(', ')
  return labels ? `Joining applies ${labels}.` : undefined
}

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

const GoToInstallButton = ({ label = 'Go to install' }: { label?: string }) => {
  const { exitToInstall } = useContext(WizardActionsContext)

  return (
    <Button type="button" variant="secondary" onClick={exitToInstall}>
      {label}
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
}) => {
  const { readOnly } = useContext(StepModeContext)
  if (readOnly) return null

  return (
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
}

const BranchEnrollment = ({
  branchId,
  groupId,
  readOnly,
  onSelect,
  onClear,
}: {
  branchId: string
  groupId: string
  readOnly: boolean
  onSelect: (branchId: string, groupId: string) => void
  onClear: () => void
}) => (
  <div className="flex flex-col gap-3 border-t pt-5">
    <div className="flex flex-wrap items-start justify-between gap-3">
      <div className="flex flex-col gap-1">
        <div className="flex items-center gap-2">
          <Text weight="strong">App branch enrollment</Text>
          <Badge size="sm">Optional</Badge>
        </div>
        <Text variant="subtext" theme="neutral">
          Join an install group now, or enroll this install later.
        </Text>
      </div>
      {!readOnly && groupId ? (
        <Button type="button" size="sm" variant="ghost" onClick={onClear}>
          Clear enrollment
        </Button>
      ) : null}
    </div>
    <div className="flex flex-col gap-2">
      {BRANCHES.map((branch) => (
        <Expand
          key={branch.id}
          id={`branch-${branch.id}`}
          isOpen={branchId ? branch.id === branchId : branch.id === BRANCHES[0].id}
          interactiveHeading
          headerToggle
          isIconBeforeHeading
          toggleLabel={`Toggle ${branch.name}`}
          headerClassName="px-3"
          className="rounded-lg border"
          heading={
            <span className="flex min-w-0 flex-1 items-center justify-between gap-3">
              <span className="flex min-w-0 items-center gap-2">
                <Icon variant="GitBranchIcon" size={16} />
                <Text weight="strong">{branch.name}</Text>
              </span>
              <Badge size="sm">
                {branch.groups.length} {branch.groups.length === 1 ? 'group' : 'groups'}
              </Badge>
            </span>
          }
        >
          <div className="flex flex-col gap-2 border-t p-3">
            {branch.groups.map((group) => {
              const description = groupDescription(group)
              return (
                <div key={group.id} className="rounded-lg bg-elevation-2 p-1">
                  <RadioInput
                    name="install-group"
                    value={group.id}
                    checked={branchId === branch.id && groupId === group.id}
                    disabled={readOnly || group.kind !== 'labels'}
                    onChange={() => onSelect(branch.id, group.id)}
                    labelProps={{
                      labelText: (
                        <span className="flex flex-wrap items-center gap-2">
                          <span>{group.name}</span>
                          {Object.entries(group.labels).map(([key, value]) => (
                            <LabelBadge key={key} labelKey={key} labelValue={value} size="sm" />
                          ))}
                        </span>
                      ),
                    }}
                  />
                  {description ? (
                    <Text variant="subtext" theme="neutral" className="px-2 pb-2">
                      {description}
                    </Text>
                  ) : null}
                </div>
              )
            })}
          </div>
        </Expand>
      ))}
    </div>
  </div>
)

const AppStep = ({ onAdvance, setSharedData, sharedData }: IWizardStepComponentProps) => {
  const { readOnly } = useContext(StepModeContext)
  const app = typeof sharedData.app === 'string' ? sharedData.app : ''
  const branch = typeof sharedData.branch === 'string' ? sharedData.branch : ''
  const installGroup = typeof sharedData.installGroup === 'string' ? sharedData.installGroup : ''

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-5">
        <Select
          labelProps={{ labelText: 'App' }}
          placeholder="Choose an app"
          searchable
          value={app}
          disabled={readOnly}
          options={APP_OPTIONS}
          onChange={(value) => {
            setSharedData('app', value)
            setSharedData('branch', '')
            setSharedData('installGroup', '')
            setSharedData(
              'stackOnly',
              APPS.find((option) => option.id === value)?.readiness === 'no-components'
            )
          }}
        />
        {app ? (
          <BranchEnrollment
            branchId={branch}
            groupId={installGroup}
            readOnly={readOnly}
            onSelect={(nextBranch, nextGroup) => {
              setSharedData('branch', nextBranch)
              setSharedData('installGroup', nextGroup)
            }}
            onClear={() => {
              setSharedData('branch', '')
              setSharedData('installGroup', '')
            }}
          />
        ) : null}
      </div>
      <StepActions
        onAdvance={onAdvance}
        label="Continue"
        showCaret
        disabled={!app}
        disabledReason="Choose an app"
      />
    </div>
  )
}

const InstallDetailsStep = ({
  onAdvance,
  onGoBack,
  setSharedData,
  sharedData,
}: IWizardStepComponentProps) => {
  const { readOnly } = useContext(StepModeContext)
  const name = typeof sharedData.name === 'string' ? sharedData.name : ''
  const region = typeof sharedData.region === 'string' ? sharedData.region : ''
  const accountId =
    typeof sharedData.accountId === 'string' ? sharedData.accountId : ''
  const autoApprove = sharedData.autoApprove === true
  const stackOnly = sharedData.stackOnly === true
  const valid = !!name.trim() && !!region && /^[0-9]{12}$/.test(accountId)

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-5">
        <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
          <Input
            id="install-name"
            labelProps={{ labelText: 'Install name' }}
            placeholder="acme-production"
            value={name}
            autoComplete="off"
            disabled={readOnly}
            onChange={(event) => setSharedData('name', event.target.value)}
          />
          <Select
            labelProps={{ labelText: 'AWS region' }}
            placeholder="Choose an AWS region"
            searchable
            value={region}
            disabled={readOnly}
            options={regionOptions('aws')}
            onChange={(value) => setSharedData('region', value)}
          />
        </div>
        <Input
          id="aws-account-id"
          labelProps={{ labelText: 'AWS account ID' }}
          helperText="The AWS account this install deploys into. It cannot be changed later."
          placeholder="000000000000"
          inputMode="numeric"
          value={accountId}
          disabled={readOnly}
          error={accountId.length > 0 && !/^[0-9]{12}$/.test(accountId)}
          errorMessage={
            accountId.length > 0 && !/^[0-9]{12}$/.test(accountId)
              ? 'Enter a 12-digit AWS account ID'
              : undefined
          }
          onChange={(event) =>
            setSharedData('accountId', event.target.value.replace(/\D/g, '').slice(0, 12))
          }
        />
        <div className="flex flex-col gap-4">
          <Toggle
            checked={autoApprove}
            disabled={readOnly}
            onChange={(checked) => setSharedData('autoApprove', checked)}
            label="Auto-approve changes"
            description="Automatically approve and apply future changes without manual confirmation."
          />
          <Toggle
            checked={stackOnly}
            disabled={readOnly}
            onChange={(checked) => setSharedData('stackOnly', checked)}
            label="Provision stack and runner only"
            description="Leave the sandbox and components unprovisioned for now."
          />
        </div>
      </div>
      <StepActions
        onGoBack={onGoBack}
        onAdvance={onAdvance}
        label="Continue"
        showCaret
        disabled={!valid}
        disabledReason="Enter the install name, region, and AWS account ID"
      />
    </div>
  )
}

const LabelsInputsStep = ({
  onAdvance,
  onGoBack,
  setSharedData,
  sharedData,
}: IWizardStepComponentProps) => {
  const { readOnly } = useContext(StepModeContext)
  const labelKey = typeof sharedData.labelKey === 'string' ? sharedData.labelKey : ''
  const labelValue =
    typeof sharedData.labelValue === 'string' ? sharedData.labelValue : ''
  const environment =
    typeof sharedData.environment === 'string' ? sharedData.environment : ''
  const replicas = typeof sharedData.replicas === 'string' ? sharedData.replicas : ''

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-3">
          <Text weight="strong">Labels</Text>
          <Text variant="subtext" theme="neutral">
            Labels organize and target this install.
          </Text>
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
            <Input
              aria-label="Label key"
              placeholder="Key"
              value={labelKey}
              disabled={readOnly}
              onChange={(event) => setSharedData('labelKey', event.target.value)}
            />
            <Input
              aria-label="Label value"
              placeholder="Value"
              value={labelValue}
              disabled={readOnly}
              onChange={(event) => setSharedData('labelValue', event.target.value)}
            />
          </div>
        </div>
        <div className="flex flex-col gap-3 border-t pt-6">
          <Text weight="strong">Inputs</Text>
          <Text variant="subtext" theme="neutral">
            Inputs configure the selected app.
          </Text>
          <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
            <Input
              id="environment-input"
              labelProps={{ labelText: 'Environment' }}
              placeholder="production"
              value={environment}
              disabled={readOnly}
              onChange={(event) => setSharedData('environment', event.target.value)}
            />
            <Input
              id="replicas-input"
              labelProps={{ labelText: 'Replicas' }}
              placeholder="3"
              inputMode="numeric"
              value={replicas}
              disabled={readOnly}
              onChange={(event) =>
                setSharedData('replicas', event.target.value.replace(/\D/g, ''))
              }
            />
          </div>
        </div>
      </div>
      <StepActions
        onGoBack={onGoBack}
        onAdvance={onAdvance}
        label="Continue"
        showCaret
      />
    </div>
  )
}

const ReviewCreateStep = ({ onAdvance, onGoBack, sharedData }: IWizardStepComponentProps) => {
  const labelKey = typeof sharedData.labelKey === 'string' ? sharedData.labelKey : ''
  const labelValue =
    typeof sharedData.labelValue === 'string' ? sharedData.labelValue : ''

  return (
    <div className="flex flex-col gap-8">
      <Card elevation="1">
        <div className="grid gap-5 sm:grid-cols-2">
          <LabeledValue label="App">{appLabel(sharedData.app) ?? 'None'}</LabeledValue>
          <LabeledValue label="Branch">
            {typeof sharedData.branch === 'string' && sharedData.branch ? (
              <Code variant="inline">{sharedData.branch}</Code>
            ) : (
              'None'
            )}
          </LabeledValue>
          <LabeledValue label="Install group">
            {groupName(sharedData.branch, sharedData.installGroup) ?? 'None'}
          </LabeledValue>
          <LabeledValue label="Install name">
            <Code variant="inline">{readName(sharedData)}</Code>
          </LabeledValue>
          <LabeledValue label="AWS region">
            {typeof sharedData.region === 'string' && sharedData.region ? (
              <CloudRegion platform="aws" region={sharedData.region} />
            ) : (
              'None'
            )}
          </LabeledValue>
          <LabeledValue label="AWS account ID">
            <Code variant="inline">{String(sharedData.accountId)}</Code>
          </LabeledValue>
          <LabeledValue label="Auto-approve changes">
            {sharedData.autoApprove === true ? 'On' : 'Off'}
          </LabeledValue>
          <LabeledValue label="Initial provisioning">
            {sharedData.stackOnly === true ? 'Stack and runner only' : 'Full install'}
          </LabeledValue>
          <LabeledValue label="Labels">
            {labelKey || labelValue ? (
              <Code variant="inline">{`${labelKey || 'key'}=${labelValue || 'value'}`}</Code>
            ) : (
              'None'
            )}
          </LabeledValue>
          <LabeledValue label="Environment">
            {typeof sharedData.environment === 'string' && sharedData.environment
              ? sharedData.environment
              : 'Not set'}
          </LabeledValue>
          <LabeledValue label="Replicas">
            {typeof sharedData.replicas === 'string' && sharedData.replicas
              ? sharedData.replicas
              : 'Not set'}
          </LabeledValue>
        </div>
        <div className="flex flex-col gap-2 border-t pt-5">
          <Text weight="strong">What will happen</Text>
          <Text theme="neutral">
            Nuon will create the install, provision its stack and runner
            {sharedData.stackOnly === true
              ? ', and stop before provisioning the sandbox and components.'
              : ', then provision the sandbox and deploy the app components.'}
          </Text>
        </div>
      </Card>
      <StepActions onGoBack={onGoBack} onAdvance={onAdvance} label="Create install" />
    </div>
  )
}

const ReviewTeardownStep = ({ onAdvance }: IWizardStepComponentProps) => (
  <div className="flex flex-col gap-8">
    <Card elevation="1">
      <div className="flex flex-col gap-2">
        <Text weight="strong">What will happen</Text>
        <Text theme="neutral">
          Deprovisioning {INSTALL_NAME} will remove it from the cloud account.
        </Text>
      </div>
      <div className="flex flex-col gap-3">
        <Text>This will create a workflow that attempts to:</Text>
        <ul className="flex list-disc flex-col gap-1 pl-6">
          <li>
            <Text theme="neutral">
              Teardown each install component according to the dependency order.
            </Text>
          </li>
          <li>
            <Text theme="neutral">Teardown the install sandbox</Text>
          </li>
        </ul>
      </div>
      <Banner theme="warn">
        <Text>
          <strong>Important:</strong> After this workflow completes, manually teardown the install
          stack.
        </Text>
      </Banner>
    </Card>
    <StepActions
      onAdvance={onAdvance}
      label="Continue"
      showCaret
      secondary={<GoToInstallButton label="Cancel teardown" />}
    />
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
        secondary={<GoToInstallButton label="Cancel teardown" />}
      />
    </div>
  )
}

const WorkflowStep = ({
  mode,
  installName,
  onFinished,
}: {
  mode: TMode
  installName: string
  onFinished?: () => void
}) => {
  const { stepMs, exitToInstall, forgetInstall } = useContext(WizardActionsContext)
  const names = mode === 'create' ? PROVISION_STEPS : TEARDOWN_STEPS
  const workflowId = mode === 'create' ? 'wf-provision' : 'wf-deprovision'
  const { progress, finished } = useSimulatedWorkflow(names, stepMs)

  useEffect(() => {
    if (finished) onFinished?.()
  }, [finished, onFinished])

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
          {mode === 'create'
            ? `${installName} is provisioned.`
            : `${installName} is removed from the cloud account. Forget the install to remove it from Nuon.`}
        </Banner>
      ) : (
        <Text theme="neutral" flex className="gap-2">
          <Icon variant="Loading" />
          {mode === 'create' ? 'Provisioning' : 'Tearing down'} {installName}
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
        {mode === 'teardown' ? (
          <GoToInstallButton label="Cancel teardown" />
        ) : !finished ? (
          <GoToInstallButton />
        ) : null}
        {mode === 'teardown' ? (
          <Button
            variant="danger"
            disabled={!finished}
            onClick={forgetInstall}
            tooltipProps={
              finished
                ? undefined
                : { tipContent: 'Wait for the teardown workflow to finish' }
            }
          >
            <Icon variant="TrashIcon" /> Forget install
          </Button>
        ) : finished ? (
          <Button variant="primary" onClick={exitToInstall}>
            View install <Icon variant="CaretRightIcon" weight="bold" />
          </Button>
        ) : null}
      </div>
    </div>
  )
}

const ProvisionStep = ({ sharedData }: IWizardStepComponentProps) => (
  <WorkflowStep mode="create" installName={readName(sharedData) || 'acme-west'} />
)

const TeardownStep = () => {
  const { markComplete } = useOnboardingWizard()
  const onFinished = useCallback(() => markComplete('teardown'), [markComplete])
  return <WorkflowStep mode="teardown" installName={INSTALL_NAME} onFinished={onFinished} />
}

const createSteps: IWizardStepDef[] = [
  {
    id: 'app',
    navLabel: 'App',
    title: 'Select app',
    description: 'Choose the app this install is created from.',
    component: AppStep,
  },
  {
    id: 'details',
    navLabel: 'Install details',
    title: 'Configure install',
    description:
      'Set the install identity, destination, approval behavior, and initial provisioning scope.',
    component: InstallDetailsStep,
  },
  {
    id: 'labels-inputs',
    navLabel: 'Labels & inputs',
    title: 'Add labels and inputs',
    description: 'Labels organize this install. Inputs configure the selected app.',
    component: LabelsInputsStep,
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
    description: 'Confirm what this teardown will remove.',
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

const EDITABLE_CREATE_STEPS = new Set(['app', 'details', 'labels-inputs'])

const SAMPLE_CREATE: Record<string, unknown> = {
  app: 'payments',
  branch: 'main',
  installGroup: 'production',
  name: 'acme-west',
  region: 'us-west-2',
  accountId: '000000000000',
}

type TCreateStepStatus = 'done' | 'current' | 'upcoming'

const CreateStepper = ({
  steps,
  selectedId,
  onSelect,
}: {
  steps: { id: string; label: string; status: TCreateStepStatus }[]
  selectedId: string
  onSelect: (id: string) => void
}) => {
  const currentIndex = steps.findIndex((step) => step.status === 'current')
  const position = currentIndex === -1 ? steps.length : currentIndex + 1

  return (
    <nav className="flex flex-col gap-3" aria-label="Setup steps">
      <ol className="flex flex-wrap items-center gap-x-2 gap-y-3">
        {steps.map((step, index) => {
          const selected = step.id === selectedId
          const reachable = step.status !== 'upcoming'
          return (
            <li key={step.id} className="flex min-w-0 items-center gap-2">
              <button
                type="button"
                disabled={!reachable}
                aria-current={step.status === 'current' ? 'step' : undefined}
                aria-pressed={selected}
                aria-label={`${index + 1}. ${step.label}`}
                onClick={() => reachable && onSelect(step.id)}
                className={cn(
                  'flex items-center gap-2 rounded-lg px-1 py-0.5 text-left outline-none',
                  'focus-visible:[--tw-outline-style:solid] focus-visible:outline-1 focus-visible:outline-offset-0 focus-visible:outline-primary-400/80',
                  reachable
                    ? 'cursor-pointer hover:bg-black/5 dark:hover:bg-white/5'
                    : 'cursor-not-allowed'
                )}
              >
                {step.status === 'done' ? (
                  <Icon
                    variant="CheckCircleIcon"
                    size={22}
                    className="text-green-600 dark:text-green-500"
                  />
                ) : (
                  <span
                    className={cn(
                      'flex size-6 shrink-0 items-center justify-center rounded-full border',
                      step.status === 'current'
                        ? 'border-primary-600 dark:border-primary-400'
                        : 'border-cool-grey-400 dark:border-dark-grey-500'
                    )}
                  >
                    <Text
                      variant="label"
                      theme={step.status === 'current' ? 'brand' : 'neutral'}
                    >
                      {index + 1}
                    </Text>
                  </span>
                )}
                <Text
                  variant="body"
                  weight={selected || step.status === 'current' ? 'strong' : 'normal'}
                  theme={
                    step.status === 'upcoming'
                      ? 'neutral'
                      : selected || step.status === 'current'
                        ? 'brand'
                        : 'default'
                  }
                >
                  {step.label}
                </Text>
              </button>
              {index === steps.length - 1 ? null : (
                <span
                  aria-hidden
                  className="h-px w-6 shrink-0 bg-cool-grey-300 dark:bg-dark-grey-600"
                />
              )}
            </li>
          )
        })}
      </ol>
      <Text variant="subtext" theme="neutral">
        Step {position} of {steps.length}
      </Text>
    </nav>
  )
}

const CreateWizard = ({
  startAtWorkflow,
  onProvision,
}: {
  startAtWorkflow: boolean
  onProvision: () => void
}) => {
  const [data, setData] = useState<Record<string, unknown>>(
    startAtWorkflow ? SAMPLE_CREATE : {}
  )
  const [done, setDone] = useState<Record<string, boolean>>(
    startAtWorkflow
      ? { app: true, details: true, 'labels-inputs': true, review: true }
      : {}
  )
  const [revisitedId, setRevisitedId] = useState<string>()
  const [editingId, setEditingId] = useState<string>()

  const currentIndex = (() => {
    const index = createSteps.findIndex((step) => step.id !== 'provision' && !done[step.id])
    return index === -1 ? createSteps.length - 1 : index
  })()
  const provisionStarted = currentIndex === createSteps.length - 1

  useEffect(() => {
    if (provisionStarted) onProvision()
  }, [provisionStarted, onProvision])

  const statuses: TCreateStepStatus[] = createSteps.map((_, index) => {
    if (index < currentIndex) return 'done'
    if (index === currentIndex) return 'current'
    return 'upcoming'
  })
  const revisitedIndex = createSteps.findIndex((step) => step.id === revisitedId)
  const revisitedStatus = revisitedIndex === -1 ? undefined : statuses[revisitedIndex]
  const showingIndex =
    revisitedStatus && revisitedStatus !== 'upcoming' ? revisitedIndex : currentIndex
  const showing = createSteps[showingIndex]
  const canEdit = !provisionStarted && EDITABLE_CREATE_STEPS.has(showing.id)
  const editing = canEdit && editingId === showing.id
  const readOnly = statuses[showingIndex] !== 'current' && !editing
  const Step = showing.component

  const setSharedData = (key: string, value: unknown) => {
    setData((current) => ({ ...current, [key]: value }))
  }

  const onAdvance = () => {
    setRevisitedId(undefined)
    setEditingId(undefined)
    if (showingIndex !== currentIndex) return
    const id = showing.id
    if (id === 'provision') return
    setDone((current) => ({ ...current, [id]: true }))
  }

  const onGoBack =
    showingIndex === currentIndex && showingIndex > 0
      ? () => {
          const previous = createSteps[showingIndex - 1]
          setRevisitedId(previous.id)
          setEditingId(EDITABLE_CREATE_STEPS.has(previous.id) ? previous.id : undefined)
        }
      : undefined

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden bg-background">
      <div className="min-h-0 flex-1 overflow-y-auto px-6 pt-8 pb-8">
        <div className="mx-auto flex w-full max-w-4xl flex-col gap-6">
          <CreateStepper
            steps={createSteps.map((step, index) => ({
              id: step.id,
              label: step.navLabel ?? step.title,
              status: statuses[index],
            }))}
            selectedId={showing.id}
            onSelect={(id) => {
              setEditingId(undefined)
              setRevisitedId(id === createSteps[currentIndex].id ? undefined : id)
            }}
          />
          <Card elevation="1">
            {revisitedStatus === 'done' && canEdit ? (
              <div className="flex justify-end">
                <Button
                  type="button"
                  size="sm"
                  variant="secondary"
                  onClick={() =>
                    setEditingId((current) => (current === showing.id ? undefined : showing.id))
                  }
                >
                  {editing ? 'Done editing' : 'Edit step'}
                </Button>
              </div>
            ) : null}
            <div className="flex flex-col gap-1">
              <Text variant="h3" role="heading" level={2}>
                {showing.title}
              </Text>
              {showing.description ? <Text theme="neutral">{showing.description}</Text> : null}
            </div>
            <StepModeContext.Provider value={{ readOnly }}>
              <Step
                isComplete={!!done[showing.id]}
                sharedData={data}
                setSharedData={setSharedData}
                onAdvance={onAdvance}
                onGoBack={onGoBack}
              />
            </StepModeContext.Provider>
          </Card>
        </div>
      </div>
    </div>
  )
}

const WizardChrome = ({
  crumbs,
  topbar,
  children,
}: {
  crumbs: TNavLink[]
  topbar?: ReactNode
  children: ReactNode
}) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(true)

  return (
    <NotificationContext.Provider value={mockNotifications}>
      <BreadcrumbContext.Provider
        value={{
          breadcrumbLinks: crumbs,
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
          <div className="flex h-full min-h-0">
            <MainLayout versions={versions}>
              <main className="flex min-h-0 w-full flex-1 flex-col">
                <MainTopbar>
                  <BreadcrumbNav />
                  {topbar}
                </MainTopbar>
                <div className="relative min-h-0 flex-1">
                  <div className="absolute inset-0">{children}</div>
                </div>
              </main>
            </MainLayout>
          </div>
        </SidebarContext.Provider>
      </BreadcrumbContext.Provider>
    </NotificationContext.Provider>
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
  const [phase, setPhase] = useState<TPhase>('wizard')
  const [session, setSession] = useState(0)
  const [cancelHidden, setCancelHidden] = useState(mode === 'create' && startAtWorkflow)
  const hideCancel = useCallback(() => setCancelHidden(true), [])

  const restart = () => {
    setSession((current) => current + 1)
    setPhase('wizard')
    setCancelHidden(mode === 'create' && startAtWorkflow)
  }

  const actions: IWizardActions = {
    stepMs,
    exitToInstall: () => setPhase('install'),
    forgetInstall: () => setPhase('forgotten'),
  }

  const result = phase === 'wizard' ? null : RESULTS[phase]
  const message = result ? (
    <PaneMessage
      title={result.title}
      description={result.description}
      actionLabel="Start over"
      onAction={restart}
    />
  ) : null

  if (mode === 'create') {
    return (
      <WizardActionsContext.Provider value={actions}>
        <WizardChrome
          crumbs={createCrumbs}
          topbar={
            phase === 'wizard' && !cancelHidden ? (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="shrink-0"
                onClick={() => setPhase('cancelled')}
              >
                Cancel
              </Button>
            ) : null
          }
        >
          {message ?? (
            <CreateWizard key={session} startAtWorkflow={startAtWorkflow} onProvision={hideCancel} />
          )}
        </WizardChrome>
      </WizardActionsContext.Provider>
    )
  }

  return (
    <WizardActionsContext.Provider value={actions}>
      <WizardChrome crumbs={teardownCrumbs}>
        <OnboardingWizardProvider
          key={session}
          steps={teardownSteps}
          initialStepIndex={startAtWorkflow ? teardownSteps.length - 1 : 0}
          onComplete={actions.exitToInstall}
        >
          {message ?? <OnboardingWizardLayout variant="pane" />}
        </OnboardingWizardProvider>
      </WizardChrome>
    </WizardActionsContext.Provider>
  )
}
