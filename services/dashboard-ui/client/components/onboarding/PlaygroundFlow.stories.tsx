export default {
  title: 'Onboarding/Playground',
}

import { createContext, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { Toggle } from '@/components/common/form/Toggle'
import {
  OnboardingWizardProvider,
  type IWizardStepComponentProps,
  type IWizardStepDef,
} from '@/providers/onboarding-wizard-provider'
import { cn } from '@/utils/classnames'
import { OnboardingWizardLayout } from './OnboardingWizard/OnboardingWizard'

const NextButton = ({
  label,
  disabled,
  disabledReason,
  onClick,
  onBack,
  showNext = true,
}: {
  label?: string
  disabled?: boolean
  disabledReason?: string
  onClick?: () => void
  onBack?: () => void
  showNext?: boolean
}) => (
  <div className={cn('flex gap-3', onBack ? 'justify-between' : 'justify-end')}>
    {onBack ? (
      <Button variant="secondary" onClick={onBack}>
        <Icon variant="CaretLeftIcon" weight="bold" /> Back
      </Button>
    ) : null}
    {showNext ? (
      <Button
        variant="primary"
        disabled={disabled}
        onClick={onClick}
        tooltipProps={disabled && disabledReason ? { tipContent: disabledReason } : undefined}
      >
        {label ?? 'Continue'} <Icon variant="CaretRightIcon" weight="bold" />
      </Button>
    ) : null}
  </div>
)

const HeroStep = ({ onAdvance, nextStepTitle }: IWizardStepComponentProps) => (
  <div className="flex flex-col gap-8 items-center text-center py-12">
    <Badge size="md" theme="brand">
      Playground
    </Badge>
    <div className="flex flex-col gap-4 max-w-xl">
      <Text variant="h1" role="heading" level={1}>
        Ship your app into your customer&apos;s cloud
      </Text>
      <Text variant="base" theme="neutral">
        Four short steps: name your org, pick your stack, choose where it runs, watch it deploy.
      </Text>
    </div>
    <Button variant="primary" size="lg" onClick={onAdvance}>
      {nextStepTitle ?? 'Get started'} <Icon variant="CaretRightIcon" weight="bold" />
    </Button>
    <Text variant="subtext" theme="neutral">
      No credit card, no cloud account needed to look around.
    </Text>
  </div>
)

const NameStep = ({
  sharedData,
  setSharedData,
  onAdvance,
  nextStepTitle,
}: IWizardStepComponentProps) => {
  const orgName = (sharedData.orgName as string) ?? ''

  return (
    <div className="flex flex-col gap-6">
      <Input
        id="playground-org-name"
        name="orgName"
        value={orgName}
        placeholder="acme-platform"
        labelProps={{ labelText: 'Organization name' }}
        onChange={(e) => setSharedData('orgName', e.currentTarget.value)}
      />
      <Text variant="subtext" theme="neutral">
        Lowercase letters, numbers and dashes. You can rename it later.
      </Text>
      <NextButton
        label={nextStepTitle}
        disabled={orgName.trim().length === 0}
        onClick={onAdvance}
      />
    </div>
  )
}

interface IChoice {
  id: string
  title: string
  description: string
  icon: TIconVariant
  badge?: string
}

const ChoiceCard = ({
  choice,
  selected,
  onSelect,
}: {
  choice: IChoice
  selected: boolean
  onSelect: () => void
}) => (
  <button
    type="button"
    onClick={onSelect}
    className={cn(
      'text-left rounded-md transition-all cursor-pointer',
      selected && 'ring-2 ring-primary-500'
    )}
  >
    <Card className={cn('h-full !gap-3', selected && '!border-primary-500')}>
      <div className="flex items-center gap-2">
        <Icon variant={choice.icon} size={20} theme={selected ? 'brand' : 'neutral'} />
        <Text variant="base" weight="strong">
          {choice.title}
        </Text>
        {choice.badge ? (
          <Badge size="sm" theme="brand">
            {choice.badge}
          </Badge>
        ) : null}
      </div>
      <Text variant="body" theme="neutral">
        {choice.description}
      </Text>
    </Card>
  </button>
)

const makeChoiceStep = (sharedKey: string, choices: IChoice[]) => {
  return function ChoiceStep({
    sharedData,
    setSharedData,
    onAdvance,
    nextStepTitle,
  }: IWizardStepComponentProps) {
    const selected = sharedData[sharedKey] as string | undefined

    return (
      <div className="flex flex-col gap-6">
        <div className="grid sm:grid-cols-2 gap-4">
          {choices.map((choice) => (
            <ChoiceCard
              key={choice.id}
              choice={choice}
              selected={selected === choice.id}
              onSelect={() => setSharedData(sharedKey, choice.id)}
            />
          ))}
        </div>
        <NextButton label={nextStepTitle} disabled={!selected} onClick={onAdvance} />
      </div>
    )
  }
}

const CLOUD_CHOICES: IChoice[] = [
  {
    id: 'sandbox',
    title: 'Nuon sandbox',
    description: 'Deploy into a managed account we run for you. Ready in a couple of minutes.',
    icon: 'FlaskIcon',
    badge: 'Fastest',
  },
  {
    id: 'aws',
    title: 'Your AWS account',
    description: 'Connect an account with an IAM role and deploy into your own VPC.',
    icon: 'CloudIcon',
  },
  {
    id: 'azure',
    title: 'Your Azure subscription',
    description: 'Bring a subscription and resource group, we handle the rest.',
    icon: 'CloudIcon',
  },
  {
    id: 'gcp',
    title: 'Your GCP project',
    description: 'Connect a project and service account to provision there.',
    icon: 'CloudIcon',
  },
]

const STACK_CHOICES: IChoice[] = [
  {
    id: 'kubernetes',
    title: 'Kubernetes + Helm',
    description: 'An EKS cluster with your Helm charts and a container registry.',
    icon: 'StackIcon',
  },
  {
    id: 'terraform',
    title: 'Terraform only',
    description: 'Straight infrastructure modules, no cluster required.',
    icon: 'CubeIcon',
  },
  {
    id: 'serverless',
    title: 'Serverless',
    description: 'Lambda functions, queues and managed databases.',
    icon: 'LightningIcon',
  },
  {
    id: 'example',
    title: 'Start from an example',
    description: 'A working sample app you can swap out later.',
    icon: 'SparkleIcon',
  },
]

const FAKE_TASKS = [
  'Creating your organization',
  'Provisioning the install sandbox',
  'Installing the Nuon runner',
  'Deploying your components',
]

const StatusStep = ({ onAdvance, nextStepTitle }: IWizardStepComponentProps) => {
  const [completed, setCompleted] = useState(0)
  const isDone = completed >= FAKE_TASKS.length

  useEffect(() => {
    if (isDone) return
    const timer = setTimeout(() => setCompleted((prev) => prev + 1), 1100)
    return () => clearTimeout(timer)
  }, [completed, isDone])

  return (
    <div className="flex flex-col gap-6">
      <Card className="!gap-0 !p-0">
        {FAKE_TASKS.map((task, index) => {
          const taskDone = index < completed
          const taskRunning = index === completed

          return (
            <div
              key={task}
              className={cn(
                'flex items-center gap-3 px-5 py-4',
                index > 0 && 'border-t'
              )}
            >
              {taskDone ? (
                <Icon variant="CheckCircleIcon" size={18} theme="success" weight="fill" />
              ) : taskRunning ? (
                <Icon variant="Loading" size={18} />
              ) : (
                <Icon variant="ClockCountdownIcon" size={18} theme="neutral" />
              )}
              <Text variant="body" theme={taskDone || taskRunning ? 'default' : 'neutral'}>
                {task}
              </Text>
              {taskDone ? (
                <Badge size="sm" theme="success" className="ml-auto">
                  Done
                </Badge>
              ) : taskRunning ? (
                <Badge size="sm" theme="info" className="ml-auto">
                  Running
                </Badge>
              ) : null}
            </div>
          )
        })}
      </Card>
      <Text variant="subtext" theme="neutral">
        {isDone
          ? 'Everything is up. This normally takes about eight minutes for a real install.'
          : 'This page updates on its own, you can leave and come back.'}
      </Text>
      <NextButton label={nextStepTitle} disabled={!isDone} onClick={onAdvance} />
    </div>
  )
}

const NEXT_LINKS: IChoice[] = [
  {
    id: 'invite',
    title: 'Invite your team',
    description: 'Add teammates and give them access to this org.',
    icon: 'UsersIcon',
  },
  {
    id: 'docs',
    title: 'Read the docs',
    description: 'Component types, inputs, actions and workflows.',
    icon: 'BookOpenTextIcon',
  },
]

const SummaryStep = ({ sharedData, onAdvance }: IWizardStepComponentProps) => {
  const orgName = (sharedData.orgName as string) || 'acme-platform'
  const cloud = (sharedData.cloud as string) || 'sandbox'
  const stack = sharedData.stack as string | undefined

  return (
    <div className="flex flex-col gap-8 py-6">
      <div className="flex flex-col gap-3 items-center text-center">
        <Icon variant="CheckCircleIcon" size={40} theme="success" weight="fill" />
        <Text variant="h2" role="heading" level={2}>
          {orgName} is live
        </Text>
        <Text variant="base" theme="neutral">
          Deployed to {cloud}
          {stack ? ` with a ${stack} stack` : ''}.
        </Text>
      </div>
      <div className="grid sm:grid-cols-2 gap-4">
        {NEXT_LINKS.map((link) => (
          <Card key={link.id} className="!gap-2">
            <div className="flex items-center gap-2">
              <Icon variant={link.icon} size={18} theme="brand" />
              <Text variant="base" weight="strong">
                {link.title}
              </Text>
            </div>
            <Text variant="body" theme="neutral">
              {link.description}
            </Text>
          </Card>
        ))}
      </div>
      <div className="flex justify-center">
        <Button variant="primary" size="lg" onClick={onAdvance}>
          Go to your dashboard
        </Button>
      </div>
    </div>
  )
}

const PlaceholderStep = ({ onAdvance, nextStepTitle }: IWizardStepComponentProps) => (
  <div className="flex flex-col gap-6">
    <Card>
      <Text variant="body" theme="neutral">
        Drop whatever you want to try here.
      </Text>
    </Card>
    <NextButton label={nextStepTitle} onClick={onAdvance} />
  </div>
)

const CloudChoiceStep = makeChoiceStep('cloud', CLOUD_CHOICES)
const StackChoiceStep = makeChoiceStep('stack', STACK_CHOICES)

const HERO_STEP: IWizardStepDef = {
  id: 'hero',
  title: 'Welcome to Nuon',
  navLabel: 'Welcome',
  hideTitle: true,
  component: HeroStep,
}

const NAME_STEP: IWizardStepDef = {
  id: 'name',
  title: 'Name your organization',
  navLabel: 'Organization',
  description: 'An org holds your apps, installs, workflows and logs.',
  component: NameStep,
}

const STACK_STEP: IWizardStepDef = {
  id: 'stack',
  title: 'Tell us about your app',
  navLabel: 'Your stack',
  description: 'Pick the shape that matches your app, or start from a working example.',
  component: StackChoiceStep,
}

const CLOUD_STEP: IWizardStepDef = {
  id: 'cloud',
  title: 'Choose where it runs',
  navLabel: 'Install',
  description: 'Use a managed sandbox to explore, or connect your own cloud account.',
  component: CloudChoiceStep,
}

const STATUS_STEP: IWizardStepDef = {
  id: 'status',
  title: 'Your install is being created',
  navLabel: 'Deploy',
  description: 'Hang tight while the resources get provisioned.',
  component: StatusStep,
}

const SUMMARY_STEP: IWizardStepDef = {
  id: 'summary',
  title: "You're all set",
  navLabel: 'Get started',
  hideTitle: true,
  component: SummaryStep,
}

const FIVE_STEP_FLOW: IWizardStepDef[] = [
  HERO_STEP,
  NAME_STEP,
  CLOUD_STEP,
  STATUS_STEP,
  SUMMARY_STEP,
]

const THREE_STEP_FLOW: IWizardStepDef[] = [HERO_STEP, CLOUD_STEP, SUMMARY_STEP]

const CHOICE_HEAVY_FLOW: IWizardStepDef[] = [
  HERO_STEP,
  STACK_STEP,
  CLOUD_STEP,
  STATUS_STEP,
  SUMMARY_STEP,
]

const MINIMAL_FLOW: IWizardStepDef[] = [
  {
    id: 'one',
    title: 'Step one',
    navLabel: 'One',
    description: 'Replace this with your own step component.',
    component: PlaceholderStep,
  },
  {
    id: 'two',
    title: 'Step two',
    navLabel: 'Two',
    description: 'Advancing past the last step finishes the flow.',
    component: PlaceholderStep,
  },
]

const FlowComplete = ({ onRestart }: { onRestart: () => void }) => (
  <div className="h-screen flex items-center justify-center bg-background">
    <Card className="items-center text-center max-w-md">
      <Icon variant="RocketIcon" size={32} theme="brand" />
      <div className="flex flex-col gap-2">
        <Text variant="h3" role="heading" level={3}>
          Flow complete
        </Text>
        <Text variant="body" theme="neutral">
          The real wizard would redirect to the dashboard here.
        </Text>
      </div>
      <Button variant="secondary" onClick={onRestart}>
        <Icon variant="ArrowCounterClockwiseIcon" size={14} /> Run it again
      </Button>
    </Card>
  </div>
)

const Playground = ({ steps }: { steps: IWizardStepDef[] }) => {
  const [runId, setRunId] = useState(0)
  const [finished, setFinished] = useState(false)

  if (finished) {
    return (
      <FlowComplete
        onRestart={() => {
          setFinished(false)
          setRunId((prev) => prev + 1)
        }}
      />
    )
  }

  return (
    <OnboardingWizardProvider
      key={runId}
      steps={steps}
      initialStepIndex={0}
      initialSharedData={{}}
      onComplete={() => setFinished(true)}
    >
      <OnboardingWizardLayout skipHref={null} />
    </OnboardingWizardProvider>
  )
}

export const FiveStep = () => <Playground steps={FIVE_STEP_FLOW} />
FiveStep.meta = { fullBleed: true }

export const ThreeStep = () => <Playground steps={THREE_STEP_FLOW} />
ThreeStep.meta = { fullBleed: true }

export const ChoiceHeavy = () => <Playground steps={CHOICE_HEAVY_FLOW} />
ChoiceHeavy.meta = { fullBleed: true }

export const Minimal = () => <Playground steps={MINIMAL_FLOW} />
Minimal.meta = { fullBleed: true }

// ---------------------------------------------------------------------------
// Fork flow: post-login screen that splits into three paths.
//   example → deploy Kitchen Sink into the user's own cloud (stack link, pre-filled inputs)
//   hosted  → deploy Kitchen Sink into an AWS account Nuon runs (no stack step at all)
//   own     → connect GitHub, set the app up from the terminal or an MCP agent
// The steps array is swapped when the fork picks a path; the provider reads
// `steps` from props on every render so the stepper follows the chosen path.
// Deliberate calls awaiting PMM review: the fork screen has no primary button
// (it is a choice screen, every path is equal weight); the dashed frames and the
// nested account → app diagram come straight from the sketch.
// ---------------------------------------------------------------------------

type TCloud = 'aws' | 'gcp' | 'azure'
type TPath = 'example' | 'hosted' | 'own'
type TForkHover = TCloud | 'sandbox' | null

interface IForkChoice {
  path: TPath
  cloud?: TCloud
}

const ForkContext = createContext<(choice: IForkChoice) => void>(() => {})
const useForkChoice = () => useContext(ForkContext)

const CLOUD_LABEL: Record<TCloud, string> = { aws: 'AWS', gcp: 'GCP', azure: 'Azure' }

const CLOUD_ICON: Record<TCloud, TIconVariant> = {
  aws: 'AWSColor',
  gcp: 'GCPColor',
  azure: 'AzureColor',
}

// Maps onto the real install stack types: aws-cloudformation, gcp-terraform, azure-bicep.
// Only AWS and Azure get a console quick link (ctl-api install_stack_version.go); GCP is applied by hand.
const CLOUD_CONNECT: Record<
  TCloud,
  { accountNoun: string; stackLabel: string; opening: string; helper: string; waitingHint: string }
> = {
  aws: {
    accountNoun: 'AWS account',
    stackLabel: 'CloudFormation stack',
    opening: 'Opening the AWS console...',
    helper:
      'Opens a pre-filled CloudFormation stack in your AWS console. Create it there, then come back — this page updates on its own.',
    waitingHint: 'Create the CloudFormation stack in the AWS console tab, then come back.',
  },
  gcp: {
    accountNoun: 'GCP project',
    stackLabel: 'Terraform stack',
    opening: 'Generating the Terraform stack...',
    helper:
      'Nuon generates a Terraform stack for your GCP project. Apply it from your terminal, then come back — this page updates on its own.',
    waitingHint: 'Apply the Terraform stack from your terminal, then come back.',
  },
  azure: {
    accountNoun: 'Azure subscription',
    stackLabel: 'Azure stack',
    opening: 'Opening the Azure portal...',
    helper:
      'Opens Deploy to Azure in the Azure portal with the stack pre-filled. Deploy it there, then come back — this page updates on its own.',
    waitingHint: 'Deploy the stack in the Azure portal tab, then come back.',
  },
}

const CLOUD_REGIONS: Record<TCloud, { label: string; options: { value: string; label: string }[] }> = {
  aws: {
    label: 'AWS region',
    options: [
      { value: 'us-east-1', label: 'us-east-1 — US East (N. Virginia)' },
      { value: 'us-west-2', label: 'us-west-2 — US West (Oregon)' },
      { value: 'eu-west-1', label: 'eu-west-1 — Europe (Ireland)' },
      { value: 'ap-southeast-1', label: 'ap-southeast-1 — Asia Pacific (Singapore)' },
    ],
  },
  gcp: {
    label: 'GCP region',
    options: [
      { value: 'us-central1', label: 'us-central1 — Iowa' },
      { value: 'us-east1', label: 'us-east1 — South Carolina' },
      { value: 'europe-west1', label: 'europe-west1 — Belgium' },
      { value: 'asia-southeast1', label: 'asia-southeast1 — Singapore' },
    ],
  },
  azure: {
    label: 'Azure location',
    options: [
      { value: 'eastus', label: 'eastus — East US' },
      { value: 'westus2', label: 'westus2 — West US 2' },
      { value: 'westeurope', label: 'westeurope — West Europe' },
      { value: 'southeastasia', label: 'southeastasia — Southeast Asia' },
    ],
  },
}

const KITCHEN_SINK_REPO = 'https://github.com/nuonco/kitchen-sink'

const readCloud = (sharedData: Record<string, unknown>): TCloud =>
  (sharedData.cloud as TCloud | undefined) ?? 'aws'

const readPath = (sharedData: Record<string, unknown>): TPath =>
  (sharedData.path as TPath | undefined) ?? 'example'

// --- Step 1: the fork --------------------------------------------------------

const INFOGRAPHIC_LABEL: Record<Exclude<TForkHover, null> | 'none', { icon: TIconVariant; text: string }> = {
  none: { icon: 'CloudIcon', text: "Your customer's cloud account" },
  aws: { icon: 'AWSColor', text: "Your customer's AWS account" },
  gcp: { icon: 'GCPColor', text: "Your customer's GCP project" },
  azure: { icon: 'AzureColor', text: "Your customer's Azure subscription" },
  sandbox: { icon: 'FlaskIcon', text: 'An AWS account Nuon runs — still yours in the CLI' },
}

const CustomerCloudInfographic = ({ hover }: { hover: TForkHover }) => {
  const label = INFOGRAPHIC_LABEL[hover ?? 'none']

  return (
    <div className="flex flex-col gap-3 rounded-lg border border-dashed p-4 bg-cool-grey-100 dark:bg-dark-grey-700">
      <div className="flex items-center gap-2">
        <Icon variant={label.icon} size={18} theme="neutral" />
        <Text variant="subtext" weight="strong" theme="neutral">
          {label.text}
        </Text>
      </div>
      <div className="flex flex-col gap-3 rounded-md border bg-background p-4">
        <div className="flex items-center justify-between gap-2">
          <Text variant="base" weight="strong">
            Your app
          </Text>
          <Badge size="sm" variant="code" theme="brand">
            kitchen-sink
          </Badge>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2">
            <Icon variant="Terraform" size={16} theme="neutral" />
            <Icon variant="Helm" size={16} theme="neutral" />
            <Icon variant="Docker" size={16} theme="neutral" />
            <Icon variant="Kubernetes" size={16} theme="neutral" />
          </div>
          <Text variant="subtext" theme="neutral">
            Terraform, Helm, images, and manifests
          </Text>
        </div>
      </div>
      <div className="flex items-center gap-2 self-end">
        <Icon variant="CpuIcon" size={14} theme="brand" />
        <Text variant="subtext" theme="neutral">
          Nuon runner — applies your components from inside the account
        </Text>
      </div>
    </div>
  )
}

const ForkStep = ({ setSharedData, onAdvance }: IWizardStepComponentProps) => {
  const choose = useForkChoice()
  const [hover, setHover] = useState<TForkHover>(null)

  const go = (choice: IForkChoice) => {
    choose(choice)
    setSharedData('path', choice.path)
    setSharedData('cloud', choice.cloud ?? 'aws')
    onAdvance()
  }

  const hoverProps = (target: TForkHover) => ({
    onMouseEnter: () => setHover(target),
    onMouseLeave: () => setHover(null),
    onFocus: () => setHover(target),
    onBlur: () => setHover(null),
  })

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <Text variant="h2" role="heading" level={2}>
          Demo how this would behave for your customer by deploying our example app
        </Text>

        <div className="grid gap-6 md:grid-cols-[1fr_1.35fr] items-center">
          <div className="flex flex-col gap-3">
            <Text variant="base" weight="strong">
              Kitchen Sink — an example app pre-wired to everything
            </Text>
            <Text variant="body" theme="neutral">
              Terraform, Helm, container images, inputs, actions, and runbooks, already configured.
              You deploy it exactly the way a customer would deploy your app.
            </Text>
            <Text variant="body" theme="neutral">
              Nothing to write.
            </Text>
          </div>
          <CustomerCloudInfographic hover={hover} />
        </div>

        <div className="flex flex-col gap-3">
          <Text variant="body" weight="strong">
            Choose a cloud to continue
          </Text>
          <div className="flex flex-wrap gap-3">
            {(['aws', 'gcp', 'azure'] as TCloud[]).map((cloud) => (
              <Button
                key={cloud}
                variant="secondary"
                size="lg"
                onClick={() => go({ path: 'example', cloud })}
                {...hoverProps(cloud)}
              >
                <Icon variant={CLOUD_ICON[cloud]} size={20} />
                Deploy to {CLOUD_LABEL[cloud]}
              </Button>
            ))}
          </div>
          <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
            <Button
              variant="secondary"
              size="lg"
              onClick={() => go({ path: 'hosted', cloud: 'aws' })}
              {...hoverProps('sandbox')}
            >
              <Icon variant="FlaskIcon" size={18} />
              Use a Nuon-hosted account
              <Badge size="sm" theme="brand">
                Fastest
              </Badge>
            </Button>
            <Text variant="subtext" theme="neutral">
              Same app, an AWS account we run. No cloud login, no stack to create.
            </Text>
          </div>
        </div>
      </Card>

      <div className="flex flex-wrap items-center justify-between gap-4 rounded-md border border-dashed p-5">
        <div className="flex items-center gap-3">
          <Icon variant="GitBranchIcon" size={24} theme="neutral" />
          <div className="flex flex-col gap-0.5">
            <Text variant="base" weight="strong">
              Already have an app?
            </Text>
            <Text variant="body" theme="neutral">
              Connect GitHub, then set it up from your terminal or your editor's agent.
            </Text>
          </div>
        </div>
        <Button variant="secondary" size="lg" onClick={() => go({ path: 'own' })}>
          Start with your app <Icon variant="CaretRightIcon" weight="bold" />
        </Button>
      </div>
    </div>
  )
}

// --- Step 2a (example path): deploy Kitchen Sink into the user's cloud ---------

type TDeployPhase = 'idle' | 'opening' | 'waiting' | 'done'

const DeployKitchenSinkStep = ({ sharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const cloud = readCloud(sharedData)
  const connect = CLOUD_CONNECT[cloud]
  const regions = CLOUD_REGIONS[cloud]

  const [phase, setPhase] = useState<TDeployPhase>('idle')
  const [region, setRegion] = useState(regions.options[0].value)
  const [autoApprove, setAutoApprove] = useState(true)

  // onAdvance is a fresh closure on every WizardStepView render; keep the timer chain stable.
  const onAdvanceRef = useRef(onAdvance)
  onAdvanceRef.current = onAdvance

  useEffect(() => {
    if (phase === 'idle') return
    const delay = phase === 'opening' ? 1200 : phase === 'waiting' ? 2600 : 900
    const timer = setTimeout(() => {
      if (phase === 'opening') setPhase('waiting')
      else if (phase === 'waiting') setPhase('done')
      else onAdvanceRef.current()
    }, delay)
    return () => clearTimeout(timer)
  }, [phase])

  const isIdle = phase === 'idle'

  const buttonLabel = isIdle
    ? `Deploy to ${CLOUD_LABEL[cloud]}`
    : phase === 'opening'
      ? connect.opening
      : phase === 'waiting'
        ? `Waiting for the ${connect.stackLabel}...`
        : `${connect.stackLabel} created`

  const statusLine =
    phase === 'waiting'
      ? `${connect.waitingHint} This page updates on its own.`
      : phase === 'done'
        ? `${connect.stackLabel} created — ${connect.accountNoun} connected`
        : null

  return (
    <div className="flex flex-col gap-6">
      <Card className="!gap-4 !p-5">
        <div className="flex flex-col gap-1">
          <Text variant="base" weight="strong">
            Install settings
          </Text>
          <Text variant="body" theme="neutral">
            Pre-selected for a quick first run. Change anything you like.
          </Text>
        </div>
        <Select
          id="fork-region"
          options={regions.options}
          labelProps={{ labelText: regions.label }}
          value={region}
          onChange={setRegion}
          disabled={!isIdle}
        />
        <Toggle
          checked={autoApprove}
          onChange={setAutoApprove}
          disabled={!isIdle}
          label="Auto-approve"
          description="Applies changes without waiting for you to approve each plan. On by default for a faster first run."
        />
      </Card>

      <Card>
        <div className="flex items-center gap-3">
          <Icon variant={CLOUD_ICON[cloud]} size={24} />
          <Text variant="base" weight="strong">
            Nuon Kitchen Sink app
          </Text>
          <Badge size="sm" variant="code">
            kitchen-sink
          </Badge>
        </div>
        <Text variant="body" theme="neutral">
          We built an example app config that has everything you can do with Nuon: Terraform modules,
          Helm charts, container images, install inputs, actions, and runbooks.{' '}
          <Link href={KITCHEN_SINK_REPO} variant="inline" isExternal>
            View app config
          </Link>
        </Text>
        <div className="flex flex-col gap-2">
          <Button variant="primary" size="lg" disabled={!isIdle} onClick={() => setPhase('opening')}>
            {buttonLabel}
            {isIdle && cloud !== 'gcp' ? <Icon variant="ArrowSquareOutIcon" size={14} /> : null}
          </Button>
          {isIdle ? (
            <Text variant="subtext" theme="neutral">
              {connect.helper}
            </Text>
          ) : null}
          {statusLine ? (
            <div className="flex items-center gap-2">
              {phase === 'done' ? (
                <Icon variant="CheckCircleIcon" size={16} theme="success" weight="fill" />
              ) : (
                <Icon variant="Loading" size={16} />
              )}
              <Text variant="body" theme={phase === 'done' ? 'success' : 'default'}>
                {statusLine}
              </Text>
            </div>
          ) : null}
        </div>
      </Card>

      {onGoBack && isIdle ? <NextButton onBack={onGoBack} showNext={false} /> : null}
    </div>
  )
}

// --- Step 2b (own path): connect GitHub, or hand it to an agent ---------------

type TGithubPhase = 'idle' | 'connecting' | 'done'

const MCP_CONFIG = `{
  "mcpServers": {
    "nuon": { "command": "nuon", "args": ["agents", "mcp", "--allow-writes"] }
  }
}`

const ConnectGithubStep = ({ onAdvance, onGoBack, nextStepTitle }: IWizardStepComponentProps) => {
  const [phase, setPhase] = useState<TGithubPhase>('idle')

  useEffect(() => {
    if (phase !== 'connecting') return
    const timer = setTimeout(() => setPhase('done'), 1600)
    return () => clearTimeout(timer)
  }, [phase])

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <div className="flex items-center gap-3">
          <Icon variant="GitHub" size={24} />
          <Text variant="base" weight="strong">
            Install the Nuon GitHub app
          </Text>
        </div>
        <Text variant="body" theme="neutral">
          Pick the repos that hold your app config and component code. You can add more later.
        </Text>
        {phase === 'done' ? (
          <div className="flex items-center gap-2">
            <Icon variant="CheckCircleIcon" size={16} theme="success" weight="fill" />
            <Text variant="body" theme="success" flex>
              Connected as
              <Badge size="sm" variant="code">
                jane-doe
              </Badge>
              — 3 repos
            </Text>
          </div>
        ) : (
          <Button
            variant="primary"
            size="lg"
            disabled={phase === 'connecting'}
            onClick={() => setPhase('connecting')}
          >
            {phase === 'connecting' ? (
              'Connecting GitHub...'
            ) : (
              <>
                <Icon variant="GitHub" size={16} />
                Connect GitHub
              </>
            )}
          </Button>
        )}
      </Card>

      <Card className="!gap-4 !p-5">
        <div className="flex items-center gap-2">
          <Icon variant="RobotIcon" size={18} theme="brand" />
          <Text variant="base" weight="strong">
            Then let your agent do the rest
          </Text>
        </div>
        <Text variant="body" theme="neutral">
          Once GitHub is connected and the CLI is installed, add the Nuon MCP server to Claude Code,
          Cursor, or any MCP client and ask it to finish setup. It can create the app, sync it, and
          create the install.
        </Text>
        <CodeBlock language="json" showCopy>
          {MCP_CONFIG}
        </CodeBlock>
        <Text variant="subtext" theme="neutral">
          Run <Badge size="sm" variant="code">nuon agents help</Badge> for the full setup guide.
        </Text>
      </Card>

      <NextButton
        label={nextStepTitle}
        onClick={onAdvance}
        onBack={onGoBack}
        disabled={phase !== 'done'}
        disabledReason="Cannot set up your app yet — connect GitHub first"
      />
    </div>
  )
}

// --- Step 3 (own path): CLI driven ---------------------------------------------

const DOCS_FIRST_APP = 'https://docs.nuon.co/get-started/create-your-first-app'

const CLI_COMMANDS: { title: string; code?: string; doc?: { href: string; text: string; note: string } }[] = [
  { title: 'Install the CLI', code: 'brew install nuonco/tap/nuon' },
  { title: 'Log in', code: 'nuon auth login' },
  {
    title: 'Write the app config',
    doc: {
      href: DOCS_FIRST_APP,
      text: 'docs.nuon.co/get-started/create-your-first-app',
      note: 'TOML in your repo: components, inputs, actions, and runbooks.',
    },
  },
  {
    title: 'Create your app from its directory, then sync it',
    code: 'nuon apps create -n my-app\nnuon apps sync',
  },
]

// The nuon-loop paste, vendored verbatim from the Kitchen Sink demo UI
// (components/ui/frontend/src/lib/nuon-loop-paste.ts → nuonco/nuon-loop PASTE.md @ a446c31, nuon-loop 0.5).
const NUON_LOOP_SPEC = 'nuon-loop 0.5'
const AGENT_PASTE =
  "/goal Set up this application on Nuon (nuon.co) so a customer can run it in their own AWS account, with no help from me unless a step truly needs a human. BOOTSTRAP FIRST, printing each result: (1) git clone --depth 1 https://github.com/nuonco/nuon-loop /tmp/nuon-loop (or gh repo clone nuonco/nuon-loop /tmp/nuon-loop if git prompts for credentials); mkdir -p .nuon-loop; copy NUON_LOOP.md and nuon-config-check.py into .nuon-loop/; read NUON_LOOP.md in full and print its Version line — it is the spec and overrides anything you assume about Nuon, and re-running it is safe. (2) Do its Phase 0 in order: locate the application (this directory if it is a git repo, else the repos/Dockerfiles/compose files/charts one level down, treated as one app unless the names clearly say otherwise — ask me only if two candidates are different products); write .claude/settings.local.json from Appendix F so nuon, python3, git and sleep never prompt me (if one still does, ask me once for \"always allow\"); make sure the nuon CLI and python3 >= 3.11 with jsonschema and pyyaml exist; nuon agents context, then nuon auth login if needed (tell me to finish the browser step), select the org if there is exactly one else ask me once, nuon apps deselect; add Appendix E to CLAUDE.md. (3) FOLLOW THE SPEC, Phases 1 through 6, under its §0.2 limits: I apply the CloudFormation stack; you never deprovision, delete or push. Infer every intake fact from the repo, its CI, its registries and this org's history; print the inferred-facts table and continue — ask me only for a fact you cannot determine and a gate depends on. THE GOAL IS MET when either (A) the transcript shows every §4 gate passing as command plus output — check script ending RESULT: PASS — 0 error(s); nuon apps validate exit 0; fresh-context review ending NO BLOCKING GAPS; nuon apps sync --no-wait with ok:true then a builds list where every component's newest build is active; an install named <app>-first (reused if it exists, else created) with an id starting inl; nuon installs stacks latest with a quick_link_url and composite_status.status awaiting-user-run; the provision workflow set to approve-all via nuon installs workflows set-approval-option (or that exact command printed in HANDOFF if you were not allowed to run it); the install's rendered readme fetched from https://api.nuon.co/v1/installs/<id>/readme returning HTTP 200 with a non-empty readme and empty warnings — and .nuon-loop/HANDOFF.md written and printed in full with no placeholders; or (B) .nuon-loop/BLOCKED.md written and printed, naming the item and gate you stopped on, every ask one that §6 allows, and the current check-script result shown. A gate counts only if its command and output are in the transcript. If I later paste a failed workflow or error, continue under Phase 7. Stop after 60 turns if neither A nor B is reached and write BLOCKED.md saying where you got stuck."
const AGENT_PASTE_PLAIN = AGENT_PASTE.replace(/^\/goal\s+/, '')

const AGENT_REQUIREMENTS = [
  'Claude Code with /goal, open in the directory that holds your app',
  'a Nuon account',
  'Python 3.11 or newer',
  'git or gh credentials with access to the private nuonco/nuon-loop',
  'AWS EKS only',
]

const CLI_EVENTS = [
  { label: 'CLI authenticated', waiting: 'Waiting for nuon auth login...' },
  { label: 'App created — my-app', waiting: 'Waiting for nuon apps create...' },
  { label: 'App synced — 3 components built', waiting: 'Waiting for nuon apps sync...' },
]

const CopyTextButton = ({
  text,
  label,
  size = 'md',
}: {
  text: string
  label: string
  size?: 'lg' | 'md' | 'sm'
}) => {
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!copied) return
    const timer = setTimeout(() => setCopied(false), 2000)
    return () => clearTimeout(timer)
  }, [copied])

  return (
    <Button
      variant="secondary"
      size={size}
      className="shrink-0"
      onClick={() => {
        navigator.clipboard?.writeText(text).catch(() => {})
        setCopied(true)
      }}
    >
      <Icon variant={copied ? 'CheckIcon' : 'CopyIcon'} size={14} />
      {copied ? 'Copied' : label}
    </Button>
  )
}

const AgentPasteCard = () => (
  <Card className="!gap-4 !p-5">
    <div className="flex flex-wrap items-center justify-between gap-2">
      <div className="flex items-center gap-2">
        <Icon variant="RobotIcon" size={18} theme="brand" />
        <Text variant="base" weight="strong">
          Let your coding agent write your Nuon app template
        </Text>
      </div>
      <Text variant="subtext" theme="neutral" family="mono">
        one paste · {NUON_LOOP_SPEC}
      </Text>
    </div>

    <div className="flex flex-col gap-4 rounded-md border p-4 sm:flex-row sm:items-center">
      <div className="line-clamp-3 flex-1">
        <Text as="span" variant="body" family="mono" weight="strong" theme="brand">
          /goal
        </Text>
        <Text as="span" variant="body" family="mono" theme="neutral">
          {AGENT_PASTE.slice(5)}
        </Text>
      </div>
      <CopyTextButton text={AGENT_PASTE} label="Copy prompt" size="lg" />
    </div>

    <div className="flex flex-wrap items-center justify-between gap-3">
      <Text variant="subtext" theme="neutral" className="max-w-xl">
        Claude Code runs it under /goal. Without /goal (Cursor, Amp) the same text runs as a plain prompt,
        with no gate evaluator.
      </Text>
      <CopyTextButton text={AGENT_PASTE_PLAIN} label="Copy without /goal" size="sm" />
    </div>

    <Expand
      id="fork-agent-requirements"
      heading={`Before you paste · ${AGENT_REQUIREMENTS.length} requirements`}
      className="border rounded-md"
    >
      <ul className="flex flex-col gap-1 p-4 pl-9 list-disc">
        {AGENT_REQUIREMENTS.map((req) => (
          <li key={req}>
            <Text variant="body" theme="neutral">
              {req}
            </Text>
          </li>
        ))}
      </ul>
    </Expand>
  </Card>
)

const CliDrivenStep = ({ onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const [completed, setCompleted] = useState(0)
  const isDone = completed >= CLI_EVENTS.length

  useEffect(() => {
    if (isDone) return
    const timer = setTimeout(() => setCompleted((prev) => prev + 1), completed === 0 ? 2600 : 1700)
    return () => clearTimeout(timer)
  }, [completed, isDone])

  return (
    <div className="flex flex-col gap-6">
      <Card className="!gap-5 !p-5">
        {CLI_COMMANDS.map((step, index) => (
          <div key={step.title} className="flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <Badge size="sm" theme="neutral">
                {index + 1}
              </Badge>
              <Text variant="body" weight="strong">
                {step.title}
              </Text>
            </div>
            {step.code ? (
              <CodeBlock language="bash" showCopy>
                {step.code}
              </CodeBlock>
            ) : null}
            {step.doc ? (
              <div className="flex flex-wrap items-center justify-between gap-3 rounded-md border px-4 py-3">
                <Link href={step.doc.href} isExternal textVariant="body">
                  {step.doc.text}
                </Link>
                <Text variant="subtext" theme="neutral">
                  {step.doc.note}
                </Text>
              </div>
            ) : null}
          </div>
        ))}
      </Card>

      <AgentPasteCard />

      <Card className="!gap-0 !p-0 overflow-hidden">
        {CLI_EVENTS.map((event, index) => {
          const eventDone = index < completed
          const eventRunning = index === completed

          return (
            <div
              key={event.label}
              className={cn('flex items-center gap-3 px-5 py-4', index > 0 && 'border-t')}
            >
              {eventDone ? (
                <Icon variant="CheckCircleIcon" size={18} theme="success" weight="fill" />
              ) : eventRunning ? (
                <Icon variant="Loading" size={18} />
              ) : (
                <Icon variant="ClockCountdownIcon" size={18} theme="neutral" />
              )}
              <Text variant="body" theme={eventDone || eventRunning ? 'default' : 'neutral'}>
                {eventDone ? event.label : event.waiting}
              </Text>
              {eventDone ? (
                <Badge size="sm" theme="success" className="ml-auto">
                  Done
                </Badge>
              ) : null}
            </div>
          )
        })}
      </Card>

      <Text variant="subtext" theme="neutral">
        {isDone
          ? 'Your app is in Nuon. Next it gets an install.'
          : 'This page updates on its own as commands finish, whether you or your agent ran them.'}
      </Text>
      <NextButton
        label="Create install"
        disabled={!isDone}
        disabledReason="Cannot create install — waiting for nuon apps sync"
        onClick={onAdvance}
        onBack={onGoBack}
      />
    </div>
  )
}

// --- Step 4 (all paths): provisioning with an explainer panel ------------------

interface IProvisionRow {
  id: string
  label: string
  copy: { pending: string; active: string; done: string }
}

// Status copy matches the live ProvisioningStep so the prototype reads like production.
const provisionRows = (path: TPath, cloud: TCloud): IProvisionRow[] => {
  const stackLabel = path === 'example' ? CLOUD_CONNECT[cloud].stackLabel : 'Install stack'
  const components = path === 'own' ? ['api', 'web', 'database'] : ['certificate', 'application_load_balancer', 'api', 'ui']

  return [
    {
      id: 'stack',
      label: stackLabel,
      copy: { pending: 'Waiting to provision...', active: 'Provisioning stack...', done: 'Stack provisioned' },
    },
    {
      id: 'runner',
      label: 'Runner',
      copy: { pending: 'Waiting to start...', active: 'Awaiting health check...', done: 'Healthy' },
    },
    {
      id: 'sandbox',
      label: 'Sandbox',
      copy: { pending: 'Waiting to configure...', active: 'Setting up your sandbox...', done: 'Sandbox ready' },
    },
    ...components.map((name) => ({
      id: `component-${name}`,
      label: name,
      copy: { pending: 'Waiting to deploy...', active: `Deploying ${name}...`, done: 'Deployed' },
    })),
  ]
}

interface IExplainerPoint {
  icon: TIconVariant
  label: string
  text: string
}

const explainerFor = (
  path: TPath,
  cloud: TCloud
): { heading: string; body: string; points: IExplainerPoint[]; cli?: string } => {
  const connect = CLOUD_CONNECT[cloud]
  const runner: IExplainerPoint = {
    icon: 'CpuIcon',
    label: 'Runner',
    text: 'A small service that boots inside the account and applies everything else.',
  }
  const sandbox: IExplainerPoint = {
    icon: 'StackIcon',
    label: 'Sandbox',
    text: 'The network and cluster your components deploy into.',
  }

  if (path === 'hosted') {
    return {
      heading: "Same install, Nuon's account",
      body: 'This install runs in an AWS account Nuon owns, but it belongs to your org. The runner, sandbox, and components are identical to a customer install — only the account is ours, so Nuon creates the stack for you.',
      points: [
        runner,
        sandbox,
        {
          icon: 'PackageIcon',
          label: 'Components',
          text: "Kitchen Sink's Terraform, Helm, and images, pulled from your org's registry.",
        },
      ],
      cli: 'nuon installs list',
    }
  }

  if (path === 'own') {
    return {
      heading: 'Your first install',
      body: 'Every customer install follows this path: stack, runner, sandbox, then your components. Watch it here or from your terminal.',
      points: [
        {
          icon: 'ShieldCheckIcon',
          label: 'Stack',
          text: 'Creates a role Nuon can assume and the network for the sandbox. The only thing a customer creates by hand.',
        },
        runner,
        sandbox,
        {
          icon: 'PackageIcon',
          label: 'Components',
          text: "Your app's Terraform, Helm, and images, built from the repos you connected.",
        },
      ],
      cli: 'nuon installs list',
    }
  }

  return {
    heading: 'This is what your customer sees',
    body: "You're running the exact install flow a customer runs. Nothing here is special to the demo.",
    points: [
      {
        icon: 'ShieldCheckIcon',
        label: 'Stack',
        text: `The ${connect.stackLabel} creates a role Nuon can assume and the network for the sandbox. It's the only thing you created by hand.`,
      },
      runner,
      sandbox,
      {
        icon: 'PackageIcon',
        label: 'Components',
        text: "Kitchen Sink's Terraform, Helm, and images, pulled from your org's registry.",
      },
    ],
  }
}

const ProvisionStep = ({ sharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const path = readPath(sharedData)
  const cloud = readCloud(sharedData)
  const rows = useMemo(() => provisionRows(path, cloud), [path, cloud])
  const explainer = explainerFor(path, cloud)

  const [completed, setCompleted] = useState(0)
  const isDone = completed >= rows.length
  const activeRow = rows[completed]
  const appName = path === 'own' ? 'my-app' : 'kitchen-sink'

  useEffect(() => {
    if (isDone) return
    const timer = setTimeout(() => setCompleted((prev) => prev + 1), 950)
    return () => clearTimeout(timer)
  }, [completed, isDone])

  const remainingMinutes = Math.max(1, Math.ceil(((rows.length - completed) * 90) / 60))

  return (
    <div className="flex flex-col gap-6">
      <div className="grid gap-6 lg:grid-cols-[3fr_2fr] items-start">
        <div className="flex flex-col gap-4">
          <Card className="!gap-0 !p-4 !flex-row items-center justify-between">
            <div className="flex items-center gap-3">
              {path === 'hosted' ? (
                <Icon variant="FlaskIcon" size={24} theme="brand" />
              ) : path === 'own' ? (
                <Icon variant="CloudIcon" size={24} theme="neutral" />
              ) : (
                <Icon variant={CLOUD_ICON[cloud]} size={24} />
              )}
              <div className="flex flex-col">
                <Text variant="base" weight="strong">
                  {appName}
                </Text>
                <Text variant="body" theme={isDone ? 'success' : 'neutral'}>
                  {isDone
                    ? 'All resources provisioned'
                    : `${activeRow.label} — ${activeRow.copy.active}`}
                </Text>
              </div>
            </div>
            {path === 'hosted' ? (
              <Badge size="sm" theme="brand">
                Nuon-hosted account
              </Badge>
            ) : (
              <Badge size="sm" theme="neutral">
                {path === 'own' ? 'Your cloud account' : CLOUD_CONNECT[cloud].accountNoun}
              </Badge>
            )}
          </Card>

          <div className="flex items-center justify-between">
            <Text variant="base" weight="strong">
              Provisioning resources
            </Text>
            <Text variant="body" theme="neutral">
              {isDone ? 'Done' : `ETA ~${remainingMinutes} min`}
            </Text>
          </div>

          <Card className="!gap-0 !p-0 overflow-hidden">
            {rows.map((row, index) => {
              const rowDone = index < completed
              const rowActive = index === completed
              const status = rowDone ? row.copy.done : rowActive ? row.copy.active : row.copy.pending

              return (
                <div
                  key={row.id}
                  className={cn('flex items-center gap-3 px-5 py-3', index > 0 && 'border-t')}
                >
                  {rowDone ? (
                    <Icon variant="CheckCircleIcon" size={18} theme="success" weight="fill" />
                  ) : rowActive ? (
                    <Icon variant="Loading" size={18} />
                  ) : (
                    <Icon variant="ClockCountdownIcon" size={18} theme="neutral" />
                  )}
                  <div className="flex flex-col">
                    <Text
                      variant="body"
                      weight="strong"
                      family={row.id.startsWith('component-') ? 'mono' : 'sans'}
                      theme={rowDone || rowActive ? 'default' : 'neutral'}
                    >
                      {row.label}
                    </Text>
                    <Text variant="subtext" theme={rowDone || rowActive ? 'success' : 'neutral'}>
                      {status}
                    </Text>
                  </div>
                </div>
              )
            })}
          </Card>
        </div>

        <Card className="!gap-4 !p-5 bg-primary-50 dark:bg-primary-950/40 !shadow-none">
          <Badge size="sm" theme="brand">
            What's happening
          </Badge>
          <div className="flex flex-col gap-2">
            <Text variant="h3" role="heading" level={3}>
              {explainer.heading}
            </Text>
            <Text variant="body" theme="neutral">
              {explainer.body}
            </Text>
          </div>
          <div className="flex flex-col gap-3">
            {explainer.points.map((point) => (
              <div key={point.label} className="flex items-start gap-3">
                <Icon variant={point.icon} size={18} theme="brand" className="mt-1 shrink-0" />
                <div className="flex flex-col">
                  <Text variant="body" weight="strong">
                    {point.label}
                  </Text>
                  <Text variant="subtext" theme="neutral">
                    {point.text}
                  </Text>
                </div>
              </div>
            ))}
          </div>
          {explainer.cli ? (
            <div className="flex flex-col gap-2">
              <Text variant="body" weight="strong">
                It's yours in the CLI too
              </Text>
              <CodeBlock language="bash" showCopy>
                {explainer.cli}
              </CodeBlock>
            </div>
          ) : null}
        </Card>
      </div>

      <Text variant="subtext" theme="neutral">
        {isDone
          ? 'Everything is up. A real install takes about eight minutes.'
          : 'This page updates on its own. You can leave and come back.'}
      </Text>
      <NextButton
        label="Finish setup"
        disabled={!isDone}
        disabledReason="Cannot finish setup — resources are still provisioning"
        onClick={onAdvance}
        onBack={onGoBack}
      />
    </div>
  )
}

// --- Step 5 (all paths): done -------------------------------------------------

const DoneStep = ({ sharedData, onAdvance }: IWizardStepComponentProps) => {
  const path = readPath(sharedData)
  const cloud = readCloud(sharedData)

  const heading =
    path === 'own'
      ? 'my-app is live'
      : path === 'hosted'
        ? 'Kitchen Sink is live in a Nuon-hosted account'
        : `Kitchen Sink is live in your ${CLOUD_CONNECT[cloud].accountNoun}`

  const body =
    path === 'own'
      ? 'Everything is provisioned and ready to go.'
      : path === 'hosted'
        ? "Everything is provisioned. Point the CLI at it, or connect your own cloud whenever you're ready."
        : 'Everything is provisioned. This is the install your customer would be looking at right now.'

  const links: IChoice[] =
    path === 'own'
      ? [
          { id: 'invite', title: 'Invite your team', description: 'Add teammates and give them access to this org.', icon: 'UsersIcon' },
          { id: 'cicd', title: 'Connect CI/CD', description: 'Trigger deploys from GitHub Actions.', icon: 'GitBranchIcon' },
        ]
      : [
          { id: 'action', title: 'Run an action', description: 'Try a runbook or action on the live install.', icon: 'PlayIcon' },
          { id: 'own', title: 'Swap in your app', description: 'Connect GitHub and deploy your own app the same way.', icon: 'GitBranchIcon' },
        ]

  return (
    <div className="flex flex-col gap-8 py-6">
      <div className="flex flex-col gap-3 items-center text-center">
        <Icon variant="CheckCircleIcon" size={40} theme="success" weight="fill" />
        <Text variant="h2" role="heading" level={2}>
          {heading}
        </Text>
        <Badge theme="success">Active</Badge>
        <Text variant="base" theme="neutral" className="max-w-xl">
          {body}
        </Text>
      </div>
      <div className="grid sm:grid-cols-2 gap-4">
        {links.map((link) => (
          <Card key={link.id} className="!gap-2 !p-4">
            <div className="flex items-center gap-2">
              <Icon variant={link.icon} size={18} theme="brand" />
              <Text variant="base" weight="strong">
                {link.title}
              </Text>
            </div>
            <Text variant="body" theme="neutral">
              {link.description}
            </Text>
          </Card>
        ))}
      </div>
      <div className="flex justify-center">
        <Button variant="primary" size="lg" onClick={onAdvance}>
          View install
        </Button>
      </div>
    </div>
  )
}

// --- Step definitions ---------------------------------------------------------

const FORK_STEP: IWizardStepDef = {
  id: 'fork',
  title: 'Welcome to Nuon',
  navLabel: 'Start',
  hideTitle: true,
  component: ForkStep,
}

const DEPLOY_STEP: Record<TCloud, IWizardStepDef> = {
  aws: {
    id: 'example-deploy-aws',
    title: 'Deploy Kitchen Sink to AWS',
    navLabel: 'Deploy',
    description: 'Create the stack in your AWS account. The install settings are pre-filled.',
    component: DeployKitchenSinkStep,
  },
  gcp: {
    id: 'example-deploy-gcp',
    title: 'Deploy Kitchen Sink to GCP',
    navLabel: 'Deploy',
    description: 'Apply the Terraform stack in your GCP project. The install settings are pre-filled.',
    component: DeployKitchenSinkStep,
  },
  azure: {
    id: 'example-deploy-azure',
    title: 'Deploy Kitchen Sink to Azure',
    navLabel: 'Deploy',
    description: 'Deploy the stack in your Azure subscription. The install settings are pre-filled.',
    component: DeployKitchenSinkStep,
  },
}

const GITHUB_STEP: IWizardStepDef = {
  id: 'own-github',
  title: 'Connect GitHub',
  navLabel: 'GitHub',
  description: 'Nuon builds your components from source. Connect the repos it should read from.',
  component: ConnectGithubStep,
}

const CLI_STEP: IWizardStepDef = {
  id: 'own-cli',
  title: 'Set up your app',
  navLabel: 'Set up',
  description: 'Four steps from your terminal, or one paste into your coding agent. This page follows along either way.',
  component: CliDrivenStep,
}

const PROVISION_STEP: Record<TPath, IWizardStepDef> = {
  example: {
    id: 'example-provision',
    title: 'Your install is being created',
    navLabel: 'Provision',
    description: 'The runner is coming up inside your account and deploying Kitchen Sink.',
    component: ProvisionStep,
  },
  hosted: {
    id: 'hosted-provision',
    title: 'Your install is being created',
    navLabel: 'Provision',
    description: 'Nuon creates the stack in an account it runs. The runner is coming up there and deploying Kitchen Sink.',
    component: ProvisionStep,
  },
  own: {
    id: 'own-provision',
    title: 'Your install is being created',
    navLabel: 'Provision',
    description: 'The runner is coming up and deploying the components you just synced.',
    component: ProvisionStep,
  },
}

const DONE_STEP: Record<TPath, IWizardStepDef> = {
  example: { id: 'example-done', title: "You're all set", navLabel: 'Done', hideTitle: true, component: DoneStep },
  hosted: { id: 'hosted-done', title: "You're all set", navLabel: 'Done', hideTitle: true, component: DoneStep },
  own: { id: 'own-done', title: "You're all set", navLabel: 'Done', hideTitle: true, component: DoneStep },
}

const buildForkFlow = (path: TPath, cloud: TCloud): IWizardStepDef[] => {
  if (path === 'hosted') return [FORK_STEP, PROVISION_STEP.hosted, DONE_STEP.hosted]
  if (path === 'own') return [FORK_STEP, GITHUB_STEP, CLI_STEP, PROVISION_STEP.own, DONE_STEP.own]
  return [
    FORK_STEP,
    DEPLOY_STEP[cloud],
    { ...PROVISION_STEP.example, id: `example-provision-${cloud}` },
    { ...DONE_STEP.example, id: `example-done-${cloud}` },
  ]
}

// --- Harness ------------------------------------------------------------------

const BranchingPlayground = ({
  initialPath = 'example',
  initialCloud = 'aws',
  initialStepIndex = 0,
}: {
  initialPath?: TPath
  initialCloud?: TCloud
  initialStepIndex?: number
}) => {
  const [runId, setRunId] = useState(0)
  const [finished, setFinished] = useState(false)
  const [choice, setChoice] = useState<IForkChoice>({ path: initialPath, cloud: initialCloud })

  const steps = useMemo(
    () => buildForkFlow(choice.path, choice.cloud ?? initialCloud),
    [choice, initialCloud]
  )

  if (finished) {
    return (
      <FlowComplete
        onRestart={() => {
          setFinished(false)
          setChoice({ path: initialPath, cloud: initialCloud })
          setRunId((prev) => prev + 1)
        }}
      />
    )
  }

  return (
    <ForkContext.Provider value={setChoice}>
      <OnboardingWizardProvider
        key={runId}
        steps={steps}
        initialStepIndex={initialStepIndex}
        initialSharedData={{ path: initialPath, cloud: initialCloud }}
        onComplete={() => setFinished(true)}
      >
        <OnboardingWizardLayout skipHref={null} />
      </OnboardingWizardProvider>
    </ForkContext.Provider>
  )
}

export const ForkFlow = () => <BranchingPlayground />
ForkFlow.meta = { fullBleed: true }

export const ForkDeployAws = () => (
  <BranchingPlayground initialPath="example" initialCloud="aws" initialStepIndex={1} />
)
ForkDeployAws.meta = { fullBleed: true }

export const ForkNuonSandbox = () => <BranchingPlayground initialPath="hosted" initialStepIndex={1} />
ForkNuonSandbox.meta = { fullBleed: true }

export const ForkOwnApp = () => <BranchingPlayground initialPath="own" initialStepIndex={1} />
ForkOwnApp.meta = { fullBleed: true }
