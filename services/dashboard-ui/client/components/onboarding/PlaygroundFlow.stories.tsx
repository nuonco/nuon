export default {
  title: 'Onboarding/Playground',
}

import { createContext, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Logo } from '@/components/common/Logo'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { Toggle } from '@/components/common/form/Toggle'
import { ToggleButton } from '@/components/common/ToggleButton'
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
// Sep 16 direction (Matt): an intro page sits BEFORE the stepper (one sentence,
// one button, a glanceable diagram), and the fork screen leads with the user's
// own app as the single primary. The example app is the secondary "kick the
// tires" path.
// ---------------------------------------------------------------------------

type TCloud = 'aws' | 'gcp' | 'azure'
type TPath = 'example' | 'hosted' | 'own'

interface IForkChoice {
  path: TPath
  cloud?: TCloud
}

// Review-only: where the template step puts the stubbed files — above the
// agent/manual tabs (editor view) or beside them (compact rows).
type TTemplateLayout = 'top' | 'beside'

interface IForkActions {
  choose: (choice: IForkChoice) => void
  backToIntro: () => void
  templateLayout: TTemplateLayout
  // Review-only: bumps each time "Simulate push" is pressed; the product's trigger is the push itself.
  pushTick: number
}

const ForkContext = createContext<IForkActions>({
  choose: () => {},
  backToIntro: () => {},
  templateLayout: 'top',
  pushTick: 0,
})
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
  {
    accountNoun: string
    stackLabel: string
    generating: string
    launch: string
    opening: string
    helper: string
    waitingHint: string
  }
> = {
  aws: {
    accountNoun: 'AWS account',
    stackLabel: 'CloudFormation stack',
    generating: 'Generating your CloudFormation stack link...',
    launch: 'Open the CloudFormation stack',
    opening: 'Opening the AWS console...',
    helper:
      'Opens a pre-filled CloudFormation stack in your AWS console. Create it there, then come back — this page updates on its own.',
    waitingHint: 'Create the CloudFormation stack in the AWS console tab, then come back.',
  },
  gcp: {
    accountNoun: 'GCP project',
    stackLabel: 'Terraform stack',
    generating: 'Generating your Terraform stack...',
    launch: 'Get the Terraform stack',
    opening: 'Preparing the Terraform stack...',
    helper:
      'Nuon generates a Terraform stack for your GCP project. Apply it from your terminal, then come back — this page updates on its own.',
    waitingHint: 'Apply the Terraform stack from your terminal, then come back.',
  },
  azure: {
    accountNoun: 'Azure subscription',
    stackLabel: 'Azure stack',
    generating: 'Generating your Azure stack link...',
    launch: 'Open the Azure stack',
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

const readAppName = (sharedData: Record<string, unknown>): string =>
  ((sharedData.appName as string | undefined) ?? '').trim() || 'my-app'

// --- Step 0: intro, outside the stepper -----------------------------------------
// Reads as an extension of sign-in: one sentence, one button, one diagram. It is
// rendered by the harness before the wizard mounts, so no stepper shows yet.

// Nuon mark without the wordmark (brand asset, not a Phosphor icon). Black in
// light mode, white in dark; same path as the design kit's logo-mark-black.svg.
const NuonMark = ({ className }: { className?: string }) => (
  <svg viewBox="0 0 23.119 32" fill="none" className={className} role="img" aria-label="Nuon">
    <path
      fill="currentColor"
      fillRule="nonzero"
      d="M 16.994 0 L 10.87 3.537 L 10.87 9.263 L 5.912 6.398 L 5.91 6.398 L 0 9.811 L 0 28.588 L 5.907 32 L 5.91 32 L 12.251 28.336 L 12.251 22.862 L 16.994 25.599 L 23.119 22.062 L 23.119 3.537 L 16.994 0 Z M 1.384 10.61 L 5.907 8 L 5.91 8 L 10.867 10.862 L 10.867 20.463 L 1.384 14.989 L 1.384 10.61 Z M 10.867 27.537 L 5.907 30.398 L 1.384 27.788 L 1.384 16.588 L 10.867 22.062 L 10.867 27.537 L 10.867 27.537 Z M 21.734 21.26 L 16.994 23.997 L 12.254 21.263 L 12.254 11.661 L 21.737 17.136 L 21.737 21.26 L 21.734 21.26 Z M 21.734 15.537 L 12.251 10.062 L 12.251 4.336 L 16.994 1.599 L 21.734 4.336 L 21.734 15.537 Z"
    />
  </svg>
)

// The same three tiers appear in both boxes: described once by the vendor,
// recreated as-is inside every customer account.
const APP_TIERS: { icon: TIconVariant; label: string }[] = [
  { icon: 'GlobeIcon', label: 'Web' },
  { icon: 'CubeIcon', label: 'API' },
  { icon: 'DatabaseIcon', label: 'Database' },
]

const MiniArch = ({ live = false }: { live?: boolean }) => (
  <div className="flex flex-wrap items-center gap-y-2">
    {APP_TIERS.map((tier, index) => (
      <div key={tier.label} className="flex items-center">
        {index > 0 ? <span className="h-px w-4 bg-neutral-200 dark:bg-neutral-600" aria-hidden /> : null}
        <div className="flex items-center gap-1.5 rounded-md border bg-background px-2.5 py-1.5">
          <Icon variant={tier.icon} size={14} theme={live ? 'brand' : 'neutral'} />
          <Text variant="subtext" weight="strong">
            {tier.label}
          </Text>
        </div>
      </div>
    ))}
  </div>
)

// What Nuon BYOC puts in the vendor's account (docs/guides/byoc.mdx and
// self-hosted.mdx — same software: dashboard-ui, ctl-api, build runner;
// Temporal/Postgres/ClickHouse behind them).
const NUON_PARTS: { icon: TIconVariant; label: string }[] = [
  { icon: 'CpuIcon', label: 'Control plane' },
  { icon: 'GlobeIcon', label: 'Dashboard' },
  { icon: 'PackageIcon', label: 'Build runner' },
]

const DEMO_REQUEST = 'https://nuon.co/demo-request'

// Collapses to zero height when closed so the layout above and below slides
// instead of jumping. grid-rows is the only height transition that needs no
// measured pixel value.
const Reveal = ({ open, children, className }: { open: boolean; children: ReactNode; className?: string }) => (
  <div
    className={cn(
      'grid transition-[grid-template-rows,opacity,transform,visibility] duration-500 ease-out',
      open ? 'visible grid-rows-[1fr] opacity-100 translate-y-0' : 'invisible grid-rows-[0fr] opacity-0 translate-y-3',
      className
    )}
    aria-hidden={!open}
  >
    <div className="overflow-hidden">{children}</div>
  </div>
)

// Levels. Hosted: your app template → Nuon (the mark on the connector) → the
// customer's account. BYOC: "Your cloud environment" grows around everything, and
// inside it a "Nuon BYOC" frame grows around the template — Nuon takes the template
// and deploys it, so the template sits inside Nuon, which sits inside your account.
const IntroDiagram = ({ selfHosted = false }: { selfHosted?: boolean }) => (
  <div className="flex flex-col gap-3">
    <div
      className={cn(
        'flex flex-col rounded-xl transition-all duration-500 ease-out',
        selfHosted
          ? 'gap-3 p-4 bg-neutral-50 dark:bg-neutral-900/60 ring-2 ring-neutral-400 dark:ring-neutral-500'
          : 'gap-0 p-0 bg-transparent ring-0 ring-transparent'
      )}
    >
      <Reveal open={selfHosted}>
        <div className="flex flex-wrap items-center justify-between gap-x-3 gap-y-2 pb-1">
          <div className="flex flex-wrap items-center gap-2">
            <Icon variant="BuildingsIcon" size={20} weight="fill" theme="neutral" />
            <Text variant="body" weight="strong">
              Your cloud environment
            </Text>
          </div>
          <div className="flex items-center gap-2">
            <Icon variant="AWSColor" size={18} />
            <Icon variant="GCPColor" size={16} />
            <Icon variant="AzureColor" size={16} />
          </div>
        </div>
      </Reveal>

      {/* Nuon BYOC frame: nothing in the hosted view, a bordered box around the
          template in BYOC. Ring, not border — the global border-color rule would
          paint a transparent border grey. */}
      <div
        className={cn(
          'flex flex-col rounded-lg transition-all duration-500 ease-out',
          selfHosted
            ? 'gap-3 p-4 bg-background ring-1 ring-neutral-200 dark:ring-neutral-700 shadow-sm'
            : 'gap-0 p-0 bg-transparent ring-1 ring-transparent shadow-none'
        )}
      >
        <Reveal open={selfHosted}>
          <div className="flex flex-col gap-3">
            <div className="flex flex-wrap items-center justify-between gap-x-3 gap-y-1">
              <div className="flex items-center gap-2">
                <NuonMark className="h-5 w-auto text-neutral-900 dark:text-white" />
                <Text variant="base" weight="strong">
                  Nuon BYOC
                </Text>
              </div>
              <Text variant="subtext" theme="neutral">
                List no subprocessors on your BYOC deals
              </Text>
            </div>
            <div className="flex flex-wrap items-center gap-y-2">
              {NUON_PARTS.map((part, index) => (
                <div key={part.label} className="flex items-center">
                  {index > 0 ? <span className="h-px w-4 bg-neutral-200 dark:bg-neutral-600" aria-hidden /> : null}
                  <div className="flex items-center gap-1.5 rounded-md border bg-background px-2.5 py-1.5">
                    <Icon variant={part.icon} size={14} theme="neutral" />
                    <Text variant="subtext" weight="strong">
                      {part.label}
                    </Text>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </Reveal>

        <div className="flex flex-col gap-3 rounded-lg border bg-background p-4 shadow-sm">
          <div className="flex items-center gap-2">
            <Icon variant="GitBranchIcon" size={18} theme="neutral" />
            <Text variant="base" weight="strong">
              Your app template
            </Text>
          </div>
          <MiniArch />
        </div>
      </div>
    </div>

    {/* The mark leaves the connector as the Nuon box above grows in: one thing moving, not two things swapping. */}
    <div className="flex items-center justify-center gap-3 py-1">
      <span
        data-intro-mark="connector"
        className={cn(
          'flex overflow-hidden transition-all duration-500 ease-out',
          selfHosted ? 'invisible w-0 -translate-y-8 scale-50 opacity-0' : 'visible w-7 translate-y-0 scale-100 opacity-100'
        )}
        aria-hidden={selfHosted}
      >
        <NuonMark className="h-7 w-auto text-neutral-900 dark:text-white" />
      </span>
      <Icon variant="ArrowDownIcon" size={24} weight="bold" theme="neutral" />
      <span
        className={cn(
          'transition-all duration-500 ease-out',
          selfHosted ? 'visible max-w-xs opacity-100' : 'invisible max-w-0 opacity-0 overflow-hidden'
        )}
        aria-hidden={!selfHosted}
      >
        <Text variant="subtext" theme="neutral" className="whitespace-nowrap">
          Deployed from your account
        </Text>
      </span>
    </div>

    {/* Offset rings behind the account stand in for "every customer". The
        account itself is the loudest thing on the page: brand tint, brand ring. */}
    <div className="relative">
      <div className="absolute inset-0 translate-x-3 translate-y-3 rounded-xl ring-2 ring-primary-200 dark:ring-primary-900" aria-hidden />
      <div className="absolute inset-0 translate-x-1.5 translate-y-1.5 rounded-xl ring-2 ring-primary-300 dark:ring-primary-800" aria-hidden />
      <div className="relative flex flex-col gap-3 rounded-xl p-4 bg-primary-50 dark:bg-primary-950/60 ring-2 ring-primary-500 shadow-md">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Icon variant="CloudIcon" size={20} weight="fill" theme="brand" />
            <Text variant="body" weight="strong">
              Your customer's cloud account
            </Text>
          </div>
          <div className="flex items-center gap-2">
            <Icon variant="AWSColor" size={18} />
            <Icon variant="GCPColor" size={16} />
            <Icon variant="AzureColor" size={16} />
          </div>
        </div>
        <div className="flex flex-col gap-3 rounded-lg border bg-background p-4 shadow-sm">
          <div className="flex items-center justify-between gap-3">
            <Text variant="base" weight="strong">
              Your app
            </Text>
            <div className="flex items-center gap-1.5">
              <span className="animate-pulse">
                <Icon variant="CheckCircleIcon" size={16} weight="fill" theme="success" />
              </span>
              <Text variant="subtext" weight="strong" theme="success">
                Running
              </Text>
            </div>
          </div>
          <MiniArch live />
        </div>
        <div className="flex items-center gap-2">
          <Icon variant="CpuIcon" size={14} theme="brand" />
          <Text variant="subtext" theme="neutral">
            Nuon runner — you operate it from here, inside their account
          </Text>
        </div>
      </div>
    </div>
  </div>
)

const IntroScreen = ({ onStart }: { onStart: () => void }) => {
  // A press button, not a link: pressed shows the BYOC picture, pressing again
  // plays the transition in reverse. The sales link lives in the panel it opens.
  const [selfHosted, setSelfHosted] = useState(false)

  return (
  <div className="h-screen flex flex-col bg-background overflow-y-auto">
    <div className="flex justify-between w-full px-6 pt-4">
      <Logo />
      <Button variant="ghost" href="https://docs.nuon.co" size="sm">
        <Icon variant="BookOpenIcon" size={14} /> Docs
      </Button>
    </div>
    <div className="flex-1 flex px-6 py-12">
      {/* my-auto centers vertically when there is room and falls back to top-aligned when the content is taller than the viewport. */}
      <div className="max-w-5xl mx-auto my-auto w-full grid gap-10 md:grid-cols-[1fr_1.2fr] items-center">
        <div className="flex flex-col gap-8">
          <div className="flex flex-col gap-3">
            <Text variant="h1" role="heading" level={1}>
              Your account is set up
            </Text>
            <Text variant="base" theme="neutral">
              The first step is to create an app template that you can deploy to a customer's cloud.
            </Text>
          </div>
          <Text variant="body" theme="neutral">
            Deploys into your customers' AWS, GCP, and Azure accounts.
          </Text>
          <div className="flex flex-col items-start gap-3">
            <Button variant="primary" size="lg" onClick={onStart}>
              Create your first app template <Icon variant="CaretRightIcon" weight="bold" />
            </Button>
            {/* Pressed = punched in: inset shadow, tint, a 1px drop. Same secondary
                variant the segmented controls use, so it reads as one family. */}
            <Button
              variant="secondary"
              size="md"
              aria-pressed={selfHosted}
              onClick={() => setSelfHosted((prev) => !prev)}
              className={cn(
                'transition-all duration-200',
                selfHosted &&
                  '!bg-primary-50 dark:!bg-primary-950/40 !shadow-[inset_0_2px_3px_rgba(0,0,0,0.14)] translate-y-px'
              )}
            >
              <Icon variant="BuildingsIcon" size={14} />
              Operating Nuon on your cloud
              <span
                className={cn('flex transition-transform duration-300', selfHosted && 'rotate-90')}
                aria-hidden
              >
                <Icon variant="CaretRightIcon" size={14} weight="bold" />
              </span>
            </Button>
          </div>
          <Reveal open={selfHosted}>
            <div className="flex flex-col gap-2 rounded-md border p-4">
              <div className="flex items-center gap-2">
                <NuonMark className="h-4 w-auto text-neutral-900 dark:text-white" />
                <Text variant="body" weight="strong">
                  Nuon BYOC
                </Text>
              </div>
              <Text variant="subtext" theme="neutral">
                The same Nuon account you are creating here, but hosted by you, managed and supported
                by us.
              </Text>
              <Link href={DEMO_REQUEST} isExternal textVariant="subtext">
                Contact sales for more
              </Link>
            </div>
          </Reveal>
        </div>
        <IntroDiagram selfHosted={selfHosted} />
      </div>
    </div>
  </div>
  )
}

// --- Step 1: the fork --------------------------------------------------------
//
// "Start with your app" does not advance the wizard. It expands the setup in
// place: connect GitHub, install the CLI, then create the app. The agent/manual
// either/or is a Clerk-style tab toggle.
//
// Order, verified against docs/guides/agents: the MCP server is `nuon agents
// mcp`, a CLI subcommand, so the CLI must be installed and logged in first. The
// nuon-loop paste drives the CLI directly and does not need the MCP server; MCP
// is offered as an optional extra. GitHub is only required for `connected_repo`
// components (private repos); production onboarding v2 has no GitHub step.

type TSetupMode = 'agent' | 'manual'

const DOCS_MCP = 'https://docs.nuon.co/guides/agents/mcp-walkthrough'
const CLI_SETUP = 'brew install nuonco/tap/nuon\nnuon auth login'
const MCP_ADD_CLAUDE = 'claude mcp add --transport stdio nuon -- nuon agents mcp --allow-writes'
const DOCS_CONFIG_FILES = 'https://docs.nuon.co/configuration-files'
const DOCS_APP_BRANCHES = 'https://docs.nuon.co/guides/app-branches'
const GIT_PUSH = (app: string) => `git add ${app}\ngit commit -m "Add Nuon app template"\ngit push origin main`
const DOCS_CONFIG_REF = 'https://docs.nuon.co/config-ref'
const DOCS_SANDBOXES = 'https://docs.nuon.co/concepts/sandboxes'
const VSCODE_EXTENSION = 'https://marketplace.visualstudio.com/items?itemName=Nuon.nuon-lsp'
// The one layout we show: config at the root of the connected repo, the way every
// example-app-config does (kitchen-sink: metadata.toml at the top level, components/
// beside it). Nesting under nuon/<app> exists in the wild but is an anti-pattern.

// The nuon-loop paste, vendored verbatim from the Kitchen Sink demo UI
// (components/ui/frontend/src/lib/nuon-loop-paste.ts → nuonco/nuon-loop PASTE.md @ a446c31, nuon-loop 0.5).
const AGENT_PASTE =
  "/goal Set up this application on Nuon (nuon.co) so a customer can run it in their own AWS account, with no help from me unless a step truly needs a human. BOOTSTRAP FIRST, printing each result: (1) git clone --depth 1 https://github.com/nuonco/nuon-loop /tmp/nuon-loop (or gh repo clone nuonco/nuon-loop /tmp/nuon-loop if git prompts for credentials); mkdir -p .nuon-loop; copy NUON_LOOP.md and nuon-config-check.py into .nuon-loop/; read NUON_LOOP.md in full and print its Version line — it is the spec and overrides anything you assume about Nuon, and re-running it is safe. (2) Do its Phase 0 in order: locate the application (this directory if it is a git repo, else the repos/Dockerfiles/compose files/charts one level down, treated as one app unless the names clearly say otherwise — ask me only if two candidates are different products); write .claude/settings.local.json from Appendix F so nuon, python3, git and sleep never prompt me (if one still does, ask me once for \"always allow\"); make sure the nuon CLI and python3 >= 3.11 with jsonschema and pyyaml exist; nuon agents context, then nuon auth login if needed (tell me to finish the browser step), select the org if there is exactly one else ask me once, nuon apps deselect; add Appendix E to CLAUDE.md. (3) FOLLOW THE SPEC, Phases 1 through 6, under its §0.2 limits: I apply the CloudFormation stack; you never deprovision, delete or push. Infer every intake fact from the repo, its CI, its registries and this org's history; print the inferred-facts table and continue — ask me only for a fact you cannot determine and a gate depends on. THE GOAL IS MET when either (A) the transcript shows every §4 gate passing as command plus output — check script ending RESULT: PASS — 0 error(s); nuon apps validate exit 0; fresh-context review ending NO BLOCKING GAPS; nuon apps sync --no-wait with ok:true then a builds list where every component's newest build is active; an install named <app>-first (reused if it exists, else created) with an id starting inl; nuon installs stacks latest with a quick_link_url and composite_status.status awaiting-user-run; the provision workflow set to approve-all via nuon installs workflows set-approval-option (or that exact command printed in HANDOFF if you were not allowed to run it); the install's rendered readme fetched from https://api.nuon.co/v1/installs/<id>/readme returning HTTP 200 with a non-empty readme and empty warnings — and .nuon-loop/HANDOFF.md written and printed in full with no placeholders; or (B) .nuon-loop/BLOCKED.md written and printed, naming the item and gate you stopped on, every ask one that §6 allows, and the current check-script result shown. A gate counts only if its command and output are in the transcript. If I later paste a failed workflow or error, continue under Phase 7. Stop after 60 turns if neither A nor B is reached and write BLOCKED.md saying where you got stuck."


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

// Collapsed preview of what "Start with your app" expands into.
const OWN_APP_STEPS: { icon: TIconVariant; title: string; body: string }[] = [
  {
    icon: 'GitHub',
    title: 'Connect GitHub',
    body: 'Needed for private repos. Nuon builds components from the repos you pick.',
  },
  {
    icon: 'RobotIcon',
    title: 'Create your app template',
    body: 'One paste into your coding agent, or write a few config files.',
  },
  {
    icon: 'CloudIcon',
    title: 'Create the first install',
    body: "Into your own cloud account first, then a customer's.",
  },
]

const AgentSetup = () => (
  <div className="flex flex-col gap-4">
    <div className="flex flex-col gap-3">
      <Text variant="body" weight="strong">
        Open the directory that holds your app, paste this to your agent
      </Text>
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
    </div>

    {/* Same shape as the sections above: strong label, bordered body, docs row. */}
    <div className="flex flex-col gap-3 rounded-md border p-4">
      <div className="flex flex-wrap items-center gap-2">
        <Badge size="sm" theme="neutral">
          Optional
        </Badge>
        <Text variant="body" weight="strong">
          Give your agent the Nuon MCP server
        </Text>
      </div>
      <Text variant="body" theme="neutral">
        Live access to your org while it works: apps, builds, installs, and logs. For example, in
        Claude Code:
      </Text>
      <CodeBlock language="bash" showCopy wrapLongLines className="!pr-14">
        {MCP_ADD_CLAUDE}
      </CodeBlock>
      <div className="flex flex-wrap items-center justify-between gap-x-3 gap-y-1">
        <Link href={DOCS_MCP} isExternal textVariant="subtext">
          docs.nuon.co/guides/agents/mcp-walkthrough
        </Link>
        <Text variant="subtext" theme="neutral">
          Cursor, Amp, and other clients take the same server as JSON. Run{' '}
          <Badge size="sm" variant="code">
            nuon agents help
          </Badge>{' '}
          for each client's file.
        </Text>
      </div>
    </div>
  </div>
)

// Manual setup, push-based. The config lives in the repo connected in Set up and the
// default app branch tracks it, so a push is the sync (docs/guides/app-branches: any
// push to the tracked branch starts a run). Three steps; the detail lives in the docs.
const ManualSetup = ({ appName, repo }: { appName: string; repo: string }) => {
  const steps: { title: string; body: ReactNode; detail?: ReactNode }[] = [
    {
      title: 'Put the config at the root of the repo you connected',
      body: (
        <>
          <Badge size="sm" variant="code">{repo}</Badge> holds the files above at its top level —{' '}
          <Badge size="sm" variant="code">metadata.toml</Badge> next to{' '}
          <Badge size="sm" variant="code">components/</Badge>, the way nuonco/kitchen-sink does. The
          directory name matches the app template name.
        </>
      ),
      detail: (
        <Link href={KITCHEN_SINK_REPO} isExternal textVariant="subtext">
          Example — nuonco/kitchen-sink
        </Link>
      ),
    },
    {
      title: 'Fill in the stubs',
      body: (
        <>
          Point each <Badge size="sm" variant="code">components/*.toml</Badge> at a repo, directory, and
          branch — a Terraform module, Helm chart, Kubernetes manifests, a container image, or a Pulumi
          program — and pick a sandbox in <Badge size="sm" variant="code">sandbox.toml</Badge>.
        </>
      ),
      detail: (
        <div className="flex flex-wrap items-center gap-x-4 gap-y-1">
          <Link href={DOCS_CONFIG_REF} isExternal textVariant="subtext">
            Component reference
          </Link>
          <Link href={DOCS_SANDBOXES} isExternal textVariant="subtext">
            Sandboxes — managed or your own
          </Link>
        </div>
      ),
    },
    {
      title: 'Commit and push',
      body: (
        <>
          Your default app branch tracks <Badge size="sm" variant="code">{repo}</Badge> on{' '}
          <Badge size="sm" variant="code">main</Badge>. Every push starts a run — no CLI step needed.
        </>
      ),
      detail: (
        <div className="flex flex-col gap-2">
          <CodeBlock language="bash" showCopy>
            {GIT_PUSH(appName)}
          </CodeBlock>
          <Text variant="subtext" theme="neutral">
            Optional: <Badge size="sm" variant="code">nuon apps validate</Badge> checks the config before you push.
          </Text>
        </div>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      <ol className="flex flex-col gap-4">
        {steps.map((step, index) => (
          <li key={step.title} className="flex gap-3">
            <Badge size="sm" theme="brand" className="mt-0.5 shrink-0">
              {index + 1}
            </Badge>
            <div className="flex min-w-0 flex-1 flex-col gap-2">
              <Text variant="body" weight="strong">
                {step.title}
              </Text>
              <Text variant="subtext" theme="neutral">
                {step.body}
              </Text>
              {step.detail}
            </div>
          </li>
        ))}
      </ol>
      <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-1 border-t pt-3">
        <div className="flex flex-wrap items-center gap-x-4 gap-y-1">
          <Link href={DOCS_APP_BRANCHES} isExternal textVariant="subtext">
            How app branches track a repo
          </Link>
          <Link href={DOCS_CONFIG_FILES} isExternal textVariant="subtext">
            Configuration files
          </Link>
        </div>
        <Text variant="subtext" theme="neutral" flex>
          Editing TOML by hand?
          <Link href={VSCODE_EXTENSION} isExternal textVariant="subtext">
            The Nuon VS Code extension
          </Link>
          adds autocomplete and inline validation.
        </Text>
      </div>
    </div>
  )
}

// The step's live moment: Nuon watching the tracked branch. In the prototype the
// review panel's "Simulate push" stands in for the push.
const PushListener = ({ repo, detected }: { repo: string; detected: boolean }) => (
  <div
    className={cn(
      'flex flex-wrap items-center justify-between gap-3 rounded-md bg-background p-4 ring-1 transition-shadow',
      detected ? 'ring-green-500 dark:ring-green-400' : 'ring-neutral-200 dark:ring-neutral-700'
    )}
  >
    <div className="flex items-center gap-3">
      {detected ? (
        <Icon variant="CheckCircleIcon" size={20} weight="fill" theme="success" />
      ) : (
        <Icon variant="Loading" size={20} />
      )}
      <div className="flex flex-col gap-0.5">
        <Text variant="body" weight="strong">
          {detected ? 'Config detected on main' : 'Listening for a push'}
        </Text>
        <Text variant="subtext" theme="neutral" flex>
          {detected ? (
            <>
              Commit
              <Badge size="sm" variant="code">
                a1b2c3d
              </Badge>
              synced the default app branch — building your components.
            </>
          ) : (
            <>
              Push to
              <Badge size="sm" variant="code">
                {repo}
              </Badge>
              or run
              <Badge size="sm" variant="code">
                nuon sync
              </Badge>
              and the default app branch picks it up.
            </>
          )}
        </Text>
      </div>
    </div>
    <Badge size="sm" theme={detected ? 'success' : 'brand'}>
      {detected ? 'Synced' : 'Watching'}
    </Badge>
  </div>
)

const SETUP_MODES: { value: TSetupMode; label: string }[] = [
  { value: 'agent', label: 'Agent setup' },
  { value: 'manual', label: 'Manual setup' },
]

// The files Nuon stubs out when the app is named. Which ones are required comes
// from the `jsonschema:"required"` tags on AppConfig (pkg/config/config.go):
// version (metadata.toml), runner, sandbox. branch.toml is the app branch Nuon
// creates behind the scenes, tracking the connected repo (docs/guides/app-branches).
// Components are where the app lives. Contents are placeholders, not a working config.
interface IAppFileStub {
  name: string
  purpose: string
  badge: string
  required: boolean
  snippet: (app: string) => string
}

const APP_FILE_STUBS: IAppFileStub[] = [
  {
    name: 'metadata.toml',
    purpose: 'Names the app template and pins the config version.',
    badge: 'Required',
    required: true,
    snippet: (app) =>
      `version      = "v1"\ndisplay_name = "${app}"\ndescription  = "What ${app} does, in one sentence."`,
  },
  {
    name: 'runner.toml',
    purpose: 'Which cloud the Nuon runner operates in.',
    badge: 'Required',
    required: true,
    snippet: () => '# aws, azure, or gcp\nrunner_type = "aws"',
  },
  {
    name: 'sandbox.toml',
    purpose: 'The base infrastructure the app lands on — EKS, AKS, GKE, ECS.',
    badge: 'Required',
    required: true,
    snippet: () =>
      'terraform_version = "1.11.3"\n\n[public_repo]\nrepo      = "nuonco/aws-eks-sandbox"\ndirectory = "."\nbranch    = "main"',
  },
  {
    name: 'branch.toml',
    purpose: 'Tracks the repo you connected — every push to main syncs the default app branch.',
    badge: 'Created for you',
    required: false,
    snippet: (app) =>
      `name = "main"\n\n[connected_repo]\nrepo      = "jane-doe/${app}"\ndirectory = "."\nbranch    = "main"`,
  },
  {
    name: 'components/api.toml',
    purpose: 'Your app: Helm charts, Terraform, images, manifests.',
    badge: 'One per component',
    required: false,
    snippet: (app) =>
      `name       = "api"\ntype       = "helm_chart"\nchart_name = "api"\nnamespace  = "${app}"\n\n[public_repo]\nrepo      = "your-org/${app}"\ndirectory = "charts/api"\nbranch    = "main"`,
  },
]

// The files were just stubbed out, so they arrive one after another on mount.
const useMountedReveal = () => {
  const [shown, setShown] = useState(false)
  useEffect(() => {
    const frame = requestAnimationFrame(() => setShown(true))
    return () => cancelAnimationFrame(frame)
  }, [])
  return shown
}
const staggerClass = (shown: boolean) =>
  cn('transition-all duration-500 ease-out', shown ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-1')
const staggerDelay = (order: number) => ({ transitionDelay: `${order * 90}ms` })

// Option A: one expandable row per file. First file open so the list reads as content, not a menu.
const FileStubRows = ({ appName }: { appName: string }) => {
  const shown = useMountedReveal()
  const [open, setOpen] = useState<string[]>([APP_FILE_STUBS[0].name])
  const toggle = (name: string) =>
    setOpen((prev) => (prev.includes(name) ? prev.filter((n) => n !== name) : [...prev, name]))

  return (
    <ul className="flex flex-col rounded-md border divide-y">
      {APP_FILE_STUBS.map((file, index) => {
        const isOpen = open.includes(file.name)
        return (
          <li key={file.name} className={cn('flex flex-col', staggerClass(shown))} style={staggerDelay(index)}>
            <button
              type="button"
              aria-expanded={isOpen}
              onClick={() => toggle(file.name)}
              className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left cursor-pointer hover:bg-cool-grey-500/8"
            >
              <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
                <Text as="span" variant="body" family="mono" weight="strong">
                  {file.name}
                </Text>
                <Badge size="sm" theme={file.required ? 'brand' : 'neutral'}>
                  {file.badge}
                </Badge>
                <Text as="span" variant="subtext" theme="neutral">
                  {file.purpose}
                </Text>
              </div>
              <span className={cn('shrink-0 transition-transform', isOpen && 'rotate-90')} aria-hidden>
                <Icon variant="CaretRightIcon" size={16} theme="neutral" />
              </span>
            </button>
            {isOpen ? (
              <div className="border-t p-3">
                <CodeBlock language="toml" wrapLongLines>
                  {file.snippet(appName)}
                </CodeBlock>
              </div>
            ) : null}
          </li>
        )
      })}
    </ul>
  )
}

// Option B: a read-only editor — file tree on the left, the selected file on the right.
const FileStubEditor = ({ appName }: { appName: string }) => {
  const shown = useMountedReveal()
  const [selected, setSelected] = useState(APP_FILE_STUBS[0].name)
  const file = APP_FILE_STUBS.find((f) => f.name === selected) ?? APP_FILE_STUBS[0]

  const treeButton = (stub: IAppFileStub, label: string, depth: 1 | 2, order: number) => (
    <button
      key={stub.name}
      type="button"
      aria-pressed={stub.name === selected}
      onClick={() => setSelected(stub.name)}
      style={staggerDelay(order)}
      className={cn(
        'flex items-center gap-2 rounded-md px-2 py-1 text-left cursor-pointer hover:bg-cool-grey-500/8',
        staggerClass(shown),
        depth === 1 ? 'ml-5' : 'ml-10',
        stub.name === selected && 'bg-primary-50 dark:bg-primary-950/40'
      )}
    >
      <Icon variant="FileCodeIcon" size={14} theme={stub.name === selected ? 'brand' : 'neutral'} />
      <Text
        as="span"
        variant="subtext"
        family="mono"
        theme={stub.name === selected ? 'brand' : 'default'}
        weight={stub.name === selected ? 'strong' : undefined}
      >
        {label}
      </Text>
    </button>
  )

  const rootFiles = APP_FILE_STUBS.filter((f) => !f.name.includes('/'))
  const componentFiles = APP_FILE_STUBS.filter((f) => f.name.startsWith('components/'))

  return (
    <div className="grid rounded-md border overflow-hidden md:grid-cols-[220px_1fr]">
      <div className="flex flex-col gap-0.5 p-2 border-b md:border-b-0 md:border-r">
        <div className={cn('flex items-center gap-2 px-2 py-1', staggerClass(shown))} style={staggerDelay(0)}>
          <Icon variant="FolderOpenIcon" size={14} theme="neutral" weight="fill" />
          <Text as="span" variant="subtext" family="mono" theme="neutral">
            {appName}/
          </Text>
        </div>
        {rootFiles.map((stub, index) => treeButton(stub, stub.name, 1, index + 1))}
        <div
          className={cn('ml-5 flex items-center gap-2 px-2 py-1', staggerClass(shown))}
          style={staggerDelay(rootFiles.length + 1)}
        >
          <Icon variant="FolderOpenIcon" size={14} theme="neutral" weight="fill" />
          <Text as="span" variant="subtext" family="mono" theme="neutral">
            components/
          </Text>
        </div>
        {componentFiles.map((stub, index) =>
          treeButton(stub, stub.name.replace('components/', ''), 2, rootFiles.length + 2 + index)
        )}
      </div>
      <div className={cn('flex min-w-0 flex-col', staggerClass(shown))} style={staggerDelay(APP_FILE_STUBS.length + 2)}>
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1 border-b px-4 py-2">
          <Text as="span" variant="body" family="mono" weight="strong">
            {file.name}
          </Text>
          <Badge size="sm" theme={file.required ? 'brand' : 'neutral'}>
            {file.badge}
          </Badge>
          <Text as="span" variant="subtext" theme="neutral">
            {file.purpose}
          </Text>
        </div>
        <CodeBlock language="toml" showLineNumbers wrapLongLines className="!rounded-none !shadow-none">
          {file.snippet(appName)}
        </CodeBlock>
      </div>
    </div>
  )
}

// The own path's escape hatch. Quiet and always in the same place, it lands back
// on the fork with the example options showing, not deep in one cloud's deploy.
const ExampleEscapeHatch = ({ onExit }: { onExit: () => void }) => (
  <div className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-dashed px-4 py-3">
    <div className="flex items-center gap-2">
      <Icon variant="PackageIcon" size={16} theme="neutral" />
      <Text variant="subtext" theme="neutral">
        Not ready to package your own app? Kick the tires with our example app instead.
      </Text>
    </div>
    <Button variant="ghost" size="sm" onClick={onExit}>
      Switch to the example app <Icon variant="ArrowRightIcon" size={14} />
    </Button>
  </div>
)

const OwnAppSetup = ({
  heading,
  appName,
  onAppName,
  githubDone,
  onGithubDone,
  showErrors,
}: {
  heading: ReactNode
  appName: string
  onAppName: (name: string) => void
  githubDone: boolean
  onGithubDone: () => void
  showErrors: boolean
}) => {
  const [connecting, setConnecting] = useState(false)
  const named = appName.trim().length > 0
  const githubError = showErrors && !githubDone

  useEffect(() => {
    if (!connecting) return
    const timer = setTimeout(() => {
      setConnecting(false)
      onGithubDone()
    }, 1600)
    return () => clearTimeout(timer)
  }, [connecting, onGithubDone])

  const githubBlock = githubDone ? (
    <Text variant="body" theme="success" flex>
      <Icon variant="CheckCircleIcon" size={18} weight="fill" theme="success" />
      Connected as
      <Badge size="sm" variant="code">
        jane-doe
      </Badge>
      — 3 repos
    </Text>
  ) : (
    <Button variant="secondary" size="lg" disabled={connecting} onClick={() => setConnecting(true)}>
      {connecting ? (
        'Connecting GitHub...'
      ) : (
        <>
          <Icon variant="GitHub" size={18} />
          Connect GitHub
        </>
      )}
    </Button>
  )

  return (
    <>
      {/* Setup: its own container, kept compact so the value card below leads. */}
      <Card className="!gap-4">
        {heading}
        <div className="grid gap-4 md:grid-cols-2">
          {/* Ring, not border, so the required/connected/error edge can change color
              (the global border-color rule would paint a border grey regardless). */}
          <div
            className={cn(
              'flex flex-col gap-3 rounded-md p-4 ring-1 transition-shadow',
              githubDone
                ? 'ring-green-500 dark:ring-green-400'
                : githubError
                  ? 'ring-red-500 dark:ring-red-400'
                  : 'ring-neutral-200 dark:ring-neutral-700'
            )}
          >
            {/* The CTA is the heading here — the button label says what this tile is. */}
            <div className="flex flex-col gap-2 items-start">
              {githubBlock}
              <Badge size="sm" theme={githubDone ? 'success' : 'brand'}>
                {githubDone ? 'Connected' : 'Required'}
              </Badge>
            </div>
            <Text variant="subtext" theme="neutral">
              Nuon integrates with your git workflow and syncs your app template.
            </Text>
            {githubError ? (
              <Text variant="subtext" theme="error" flex>
                <Icon variant="WarningCircleIcon" size={14} weight="fill" />
                Connect GitHub to continue.
              </Text>
            ) : null}
          </div>
          <div className="flex flex-col gap-3 rounded-md border p-4">
            <div className="flex items-center gap-2">
              <Icon variant="TerminalWindowIcon" size={20} theme="neutral" />
              <Text variant="body" weight="strong">
                Install the CLI and log in
              </Text>
            </div>
            <CodeBlock language="bash" showCopy>
              {CLI_SETUP}
            </CodeBlock>
          </div>
        </div>
      </Card>

      {/* Name your app: typing a name is the moment the app starts to exist. The files
          it needs are stubbed out on the next step, where they get filled in. */}
      <Card className="!gap-4">
        <div className="flex flex-col gap-1">
          <Text variant="h3" role="heading" level={3}>
            Name your app template
          </Text>
          <Text variant="body" theme="neutral">
            Nuon creates the app template and stubs out the config files it needs.
          </Text>
        </div>
        <div className="max-w-sm">
          <Input
            id="fork-app-name"
            size="lg"
            placeholder="my-app"
            value={appName}
            onChange={(e) => onAppName(e.currentTarget.value)}
            labelProps={{ labelText: 'App template name' }}
            helperText="Also the name of the directory that holds its config."
            error={showErrors && !named}
            errorMessage="Name your app template to continue."
            autoComplete="off"
            spellCheck={false}
          />
        </div>
      </Card>
    </>
  )
}

// --- Step 1b: the template (own path only) ------------------------------------
//
// The app exists and its config is stubbed. This step shows the stubs and the
// two ways to fill them in. Where the stubs sit is a review toggle: above the
// tabs as a read-only editor, or beside them as compact rows.
const TemplateStep = ({ sharedData, setSharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const { templateLayout, choose, pushTick } = useForkChoice()
  const appName = readAppName(sharedData)
  const [mode, setMode] = useState<TSetupMode>('agent')
  const beside = templateLayout === 'beside'
  // The connected account from Set up, and a repo named after the template.
  const repo = `jane-doe/${appName}`
  const detected = pushTick > 0

  // Back to the fork, collapsed, with the example path selected.
  const exitToExample = () => {
    setSharedData('expandOwn', false)
    setSharedData('path', 'example')
    choose({ path: 'example', cloud: readCloud(sharedData) })
    onGoBack?.()
  }

  const stubs = (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center gap-2">
        <Text variant="body" weight="strong">
          Your app template
        </Text>
        <Badge size="sm" theme="brand">
          Stubbed by Nuon
        </Badge>
        <Text variant="subtext" theme="neutral">
          Three required files, the branch that tracks your repo, and components/.
        </Text>
      </div>
      {beside ? <FileStubRows appName={appName} /> : <FileStubEditor appName={appName} />}
    </div>
  )

  const fill = (
    <div className="flex flex-col gap-3">
      <Text variant="body" weight="strong">
        Fill it in
      </Text>
      <ToggleButton<TSetupMode>
        options={SETUP_MODES}
        value={mode}
        onChange={setMode}
        size="lg"
        className="self-start"
      />
      <div className="rounded-md border bg-background p-4">
        {mode === 'agent' ? <AgentSetup /> : <ManualSetup appName={appName} repo={repo} />}
      </div>
    </div>
  )

  return (
    <div className="flex flex-col gap-6">
      <Card className="!gap-5 !p-5 !border-0 !shadow-none bg-primary-50 dark:bg-primary-950/40 ring-1 ring-primary-200 dark:ring-primary-800">
        <Text variant="body" theme="neutral" flex>
          You are creating
          <Badge size="sm" variant="code">
            {appName}
          </Badge>
        </Text>
        {beside ? (
          <div className="grid gap-5 items-start lg:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)]">
            {stubs}
            {fill}
          </div>
        ) : (
          <>
            {stubs}
            {fill}
          </>
        )}
        <PushListener repo={repo} detected={detected} />
      </Card>
      <ExampleEscapeHatch onExit={exitToExample} />
      <NextButton
        label="Set up your first install"
        disabled={!detected}
        disabledReason="Cannot continue — waiting for your first push"
        onClick={onAdvance}
        onBack={onGoBack}
      />
    </div>
  )
}

const ForkStep = ({ sharedData, setSharedData, onAdvance }: IWizardStepComponentProps) => {
  const { choose, backToIntro } = useForkChoice()
  const [expanded, setExpanded] = useState(Boolean(sharedData.expandOwn))
  const setupRef = useRef<HTMLDivElement>(null)
  const named = ((sharedData.appName as string | undefined) ?? '').trim().length > 0
  const githubDone = Boolean(sharedData.githubDone)
  // Errors show only after a failed attempt to continue, on whichever field is missing.
  const [showErrors, setShowErrors] = useState(false)

  const tryContinue = () => {
    if (!named || !githubDone) {
      setShowErrors(true)
      setupRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
      return
    }
    go({ path: 'own' })
  }

  const go = (choice: IForkChoice) => {
    choose(choice)
    setSharedData('path', choice.path)
    setSharedData('cloud', choice.cloud ?? 'aws')
    onAdvance()
  }

  // Expanding commits to the own-app path so the stepper stops showing the example path's "Deploy" dot.
  const expand = () => {
    choose({ path: 'own' })
    setSharedData('path', 'own')
    // Persisted so Back from the template step remounts this step still expanded.
    setSharedData('expandOwn', true)
    setExpanded(true)
    requestAnimationFrame(() => setupRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
  }

  const exitToExample = () => {
    setSharedData('expandOwn', false)
    setSharedData('path', 'example')
    choose({ path: 'example', cloud: readCloud(sharedData) })
    setExpanded(false)
  }

  const heading = (
    <Text variant="h2" role="heading" level={2}>
      Deploy your app to a customer's cloud
    </Text>
  )
  const setupHeading = (
    <Text variant="h2" role="heading" level={2}>
      Set up
    </Text>
  )

  return (
    <div className="flex flex-col gap-6">
      <Text variant="h1" role="heading" level={1}>
        Create your first app template
      </Text>
      {expanded ? (
        <div ref={setupRef} className="flex flex-col gap-6 scroll-mt-6">
          <OwnAppSetup
            heading={setupHeading}
            appName={(sharedData.appName as string | undefined) ?? ''}
            onAppName={(name) => setSharedData('appName', name)}
            githubDone={githubDone}
            onGithubDone={() => setSharedData('githubDone', true)}
            showErrors={showErrors}
          />
          <ExampleEscapeHatch onExit={exitToExample} />
        </div>
      ) : (
        <>
        <Card>
          {heading}
          <ol className="grid gap-3 sm:grid-cols-3">
              {OWN_APP_STEPS.map((step, index) => (
                <li key={step.title} className="flex flex-col gap-2 rounded-md border p-4">
                  <div className="flex items-center justify-between gap-2">
                    <Icon variant={step.icon} size={20} theme="brand" />
                    <Badge size="sm" theme="brand">
                      {index + 1}
                    </Badge>
                  </div>
                  <Text variant="body" weight="strong">
                    {step.title}
                  </Text>
                  <Text variant="subtext" theme="neutral">
                    {step.body}
                  </Text>
                </li>
              ))}
            </ol>
          <div>
            <Button variant="primary" size="lg" onClick={expand}>
              Start with your app <Icon variant="CaretRightIcon" weight="bold" />
            </Button>
          </div>
        </Card>
      <Card className="!gap-4">
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-2">
            <Icon variant="PackageIcon" size={20} theme="neutral" />
            <Text variant="h3" role="heading" level={3}>
              Or kick the tires with our example app first
            </Text>
          </div>
          <Text variant="body" theme="neutral">
            Pre-wired with Terraform, Helm, images, and manifests, and deploys exactly the way a
            customer would deploy yours.
          </Text>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          {(['aws', 'gcp', 'azure'] as TCloud[]).map((cloud) => (
            <Button
              key={cloud}
              variant="secondary"
              size="md"
              onClick={() => go({ path: 'example', cloud })}
            >
              <Icon variant={CLOUD_ICON[cloud]} size={18} />
              Deploy to {CLOUD_LABEL[cloud]}
            </Button>
          ))}
          <Button variant="secondary" size="md" onClick={() => go({ path: 'hosted', cloud: 'aws' })}>
            <Icon variant="FlaskIcon" size={16} />
            Use a Nuon-hosted account
            <Badge size="sm" theme="brand">
              Fastest
            </Badge>
          </Button>
        </div>
        <Text variant="subtext" theme="neutral">
          Nuon-hosted. Great way to test out the CLI and product on a real example app — without
          incurring your own POC cloud costs.
        </Text>
      </Card>
        </>
      )}


      {expanded ? (
        <NextButton label="Create your app template" onClick={tryContinue} onBack={backToIntro} />
      ) : (
        <NextButton onBack={backToIntro} showNext={false} />
      )}
    </div>
  )
}

// --- Step 2a (example path): deploy Kitchen Sink into the user's cloud ---------

// One phase machine for the stack: Nuon renders the link (about 30s in the product),
// the customer launches it, the stack reports back.
type TStackPhase = 'generating' | 'ready' | 'opening' | 'waiting' | 'done'

const CLOUD_OPTIONS: { value: TCloud; label: string }[] = [
  { value: 'aws', label: 'AWS' },
  { value: 'gcp', label: 'GCP' },
  { value: 'azure', label: 'Azure' },
]

// How a customer creates the install stack, per cloud (docs/concepts/stacks.mdx and
// docs/platform-support/*): Terraform plus the platform's native format, except GCP,
// which is Terraform only. The first entry is the one this install's link uses.
const DOCS_STACKS = 'https://docs.nuon.co/concepts/stacks'
const STACK_METHODS: Record<TCloud, { name: string; how: string }[]> = {
  aws: [
    { name: 'CloudFormation quick-create', how: 'One pre-filled link. Your customer creates the stack in their console.' },
    { name: 'AWS CLI', how: 'The same CloudFormation template, from a terminal.' },
    { name: 'Terraform', how: 'Generated tfvars for the install-stacks/aws module, applied with terraform.' },
  ],
  gcp: [
    { name: 'Terraform', how: 'Generated tfvars for the install-stacks/gcp module. gcloud auth, then terraform apply. GCP is Terraform only.' },
  ],
  azure: [
    { name: 'Deploy to Azure (Bicep)', how: 'One pre-filled link. Your customer deploys in the portal.' },
    { name: 'Terraform', how: 'Generated tfvars for the install-stacks/azure module, applied with terraform.' },
  ],
}

// What this install will contain, as a card worth reading: source, sandbox, and
// components, plus the same three tiers the intro drew. Framing-agnostic — the
// example and own paths differ only in the facts.
const InstallSummaryCard = ({ path, appName }: { path: TPath; appName: string }) => {
  const own = path === 'own'
  const facts: { label: string; value: ReactNode }[] = own
    ? [
        {
          label: 'Source',
          value: (
            <>
              <Badge size="sm" variant="code">
                jane-doe/{appName}
              </Badge>
              <Badge size="sm" variant="code">
                main
              </Badge>
            </>
          ),
        },
        { label: 'Sandbox', value: 'Nuon-managed, from sandbox.toml' },
        { label: 'Components', value: 'api — Helm chart, from components/api.toml' },
      ]
    : [
        {
          label: 'Source',
          value: (
            <Badge size="sm" variant="code">
              nuonco/kitchen-sink
            </Badge>
          ),
        },
        { label: 'Sandbox', value: 'Nuon-managed EKS sandbox' },
        { label: 'Components', value: 'Terraform modules, Helm charts, container images' },
      ]

  return (
    <Card className="!gap-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Icon variant={own ? 'GitBranchIcon' : 'PackageIcon'} size={24} theme="brand" />
          <div className="flex flex-col">
            <Text variant="base" weight="strong">
              {own ? appName : 'Nuon Kitchen Sink app'}
            </Text>
            <Text variant="subtext" theme="neutral">
              {own
                ? 'Your app template. This is what every install of it will contain.'
                : 'Everything you can do with Nuon, in one example app.'}
            </Text>
          </div>
        </div>
        {own ? null : (
          <Link href={KITCHEN_SINK_REPO} isExternal textVariant="subtext">
            View app config
          </Link>
        )}
      </div>
      <dl className="grid gap-4 sm:grid-cols-3">
        {facts.map((fact) => (
          <div key={fact.label} className="flex flex-col gap-1.5 rounded-md border p-3">
            <dt>
              <Text variant="subtext" theme="neutral">
                {fact.label}
              </Text>
            </dt>
            <dd className="flex flex-wrap items-center gap-1.5">
              {typeof fact.value === 'string' ? (
                <Text variant="body" weight="strong">
                  {fact.value}
                </Text>
              ) : (
                fact.value
              )}
            </dd>
          </div>
        ))}
      </dl>
      <div className="flex flex-col gap-2">
        <Text variant="subtext" theme="neutral">
          What lands in the customer's account
        </Text>
        <MiniArch />
      </div>
    </Card>
  )
}

// --- Step: set up the install --------------------------------------------------
//
// Settings and the app, nothing else. "Create install" is the moment the install
// exists; the stack link starts generating on the next step.
const DeployStep = ({ sharedData, setSharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const path = readPath(sharedData)
  const cloud = readCloud(sharedData)
  const regions = CLOUD_REGIONS[cloud]
  const appName = path === 'own' ? readAppName(sharedData) : 'Kitchen Sink'
  const region = (sharedData.region as string | undefined) ?? regions.options[0].value
  const [autoApprove, setAutoApprove] = useState(true)

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
        {path === 'own' ? (
          <div className="flex flex-col gap-2">
            <Text variant="body" weight="strong">
              Cloud
            </Text>
            <ToggleButton<TCloud>
              options={CLOUD_OPTIONS}
              value={cloud}
              onChange={(value) => {
                setSharedData('cloud', value)
                setSharedData('region', CLOUD_REGIONS[value].options[0].value)
              }}
              size="md"
              className="self-start"
            />
          </div>
        ) : null}
        <Select
          id="fork-region"
          options={regions.options}
          labelProps={{ labelText: regions.label }}
          value={region}
          onChange={(value) => setSharedData('region', value)}
        />
        <Toggle
          checked={autoApprove}
          onChange={setAutoApprove}
          label="Auto-approve"
          description="Applies changes without waiting for you to approve each plan. On by default for a faster first run."
        />
      </Card>

      <InstallSummaryCard path={path} appName={appName} />

      <NextButton label="Create install" onClick={onAdvance} onBack={onGoBack} />
    </div>
  )
}

// --- Step: the install stack ----------------------------------------------------
//
// Two things happen here. While Nuon renders the stack link (about 30s), the
// page explains how a customer creates that stack — quick-create, CLI, Terraform.
// When the link exists the primary becomes the link itself; once the stack
// reports back this advances to the workflow explainer. Users never get past
// this page before the link exists.
const StackStep = ({ sharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const path = readPath(sharedData)
  const cloud = readCloud(sharedData)
  const connect = CLOUD_CONNECT[cloud]
  const regions = CLOUD_REGIONS[cloud]
  const appName = path === 'own' ? readAppName(sharedData) : 'Kitchen Sink'
  const region = (sharedData.region as string | undefined) ?? regions.options[0].value
  const regionLabel = regions.options.find((option) => option.value === region)?.label ?? region
  const methods = STACK_METHODS[cloud]

  const [phase, setPhase] = useState<TStackPhase>('generating')
  const onAdvanceRef = useRef(onAdvance)
  onAdvanceRef.current = onAdvance

  useEffect(() => {
    if (phase === 'ready') return
    const delay =
      phase === 'generating' ? 3200 : phase === 'opening' ? 1200 : phase === 'waiting' ? 2600 : 900
    const timer = setTimeout(() => {
      if (phase === 'generating') setPhase('ready')
      else if (phase === 'opening') setPhase('waiting')
      else if (phase === 'waiting') setPhase('done')
      else onAdvanceRef.current()
    }, delay)
    return () => clearTimeout(timer)
  }, [phase])

  const generating = phase === 'generating'
  const ready = phase === 'ready'
  const beforeLaunch = generating || ready

  const buttonLabel = generating
    ? connect.generating
    : ready
      ? connect.launch
      : phase === 'opening'
        ? connect.opening
        : phase === 'waiting'
          ? `Waiting for the ${connect.stackLabel}...`
          : `${connect.stackLabel} created`

  const status = generating
    ? `Generating the ${connect.stackLabel} link for ${regionLabel}. About 30 seconds.`
    : ready
      ? `${connect.stackLabel} link ready for ${region}. From launch to a healthy runner is about 11 minutes — this page updates on its own.`
      : phase === 'waiting'
        ? `${connect.waitingHint} This page updates on its own.`
        : phase === 'done'
          ? `${connect.stackLabel} created — ${connect.accountNoun} connected.`
          : connect.opening

  return (
    <div className="flex flex-col gap-6">
      <Card className="!gap-0 !p-4 !flex-row items-center justify-between">
        <div className="flex items-center gap-3">
          <Icon variant={CLOUD_ICON[cloud]} size={24} />
          <div className="flex flex-col">
            <Text variant="base" weight="strong">
              {connect.stackLabel} for {appName}
            </Text>
            <Text variant="body" theme="neutral">
              {connect.accountNoun} · {region}
            </Text>
          </div>
        </div>
        <Badge size="sm" theme={ready || phase === 'done' ? 'success' : 'brand'}>
          {generating ? 'Generating' : ready ? 'Ready' : phase === 'done' ? 'Created' : 'Waiting'}
        </Badge>
      </Card>

      <Card className="!gap-5">
        <div className="flex flex-col gap-1">
          <Text variant="h3" role="heading" level={3}>
            How your customers create this install
          </Text>
          <Text variant="body" theme="neutral">
            Nuon renders the install stack in Terraform and in {CLOUD_LABEL[cloud]}'s native format —
            same resources either way. Your customer creates it with their own credentials; that is how
            access is granted. You are about to do it the way they would.
          </Text>
        </div>
        <ul className="flex flex-col divide-y rounded-md border">
          {methods.map((method, index) => (
            <li key={method.name} className="flex flex-wrap items-baseline gap-x-3 gap-y-1 px-4 py-3">
              <Text variant="body" weight="strong">
                {method.name}
              </Text>
              {index === 0 ? (
                <Badge size="sm" theme="brand">
                  This install
                </Badge>
              ) : null}
              <Text variant="subtext" theme="neutral">
                {method.how}
              </Text>
            </li>
          ))}
        </ul>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            {ready || phase === 'done' ? (
              <Icon variant="CheckCircleIcon" size={16} theme="success" weight="fill" />
            ) : (
              <Icon variant="Loading" size={16} />
            )}
            <Text variant="subtext" theme={ready || phase === 'done' ? 'success' : 'neutral'}>
              {status}
            </Text>
          </div>
          <Link href={DOCS_STACKS} isExternal textVariant="subtext">
            All formats and CLI snippets
          </Link>
        </div>
      </Card>

      <div className="flex items-center justify-between gap-3">
        {onGoBack && beforeLaunch ? (
          <Button variant="secondary" size="lg" onClick={onGoBack}>
            <Icon variant="CaretLeftIcon" weight="bold" /> Back
          </Button>
        ) : (
          <span />
        )}
        <Button
          variant="primary"
          size="lg"
          disabled={!ready}
          onClick={() => setPhase('opening')}
          tooltipProps={
            generating ? { tipContent: `Cannot launch yet — Nuon is still generating the ${connect.stackLabel} link` } : undefined
          }
        >
          {generating || phase === 'opening' || phase === 'waiting' ? <Icon variant="Loading" size={16} /> : null}
          {buttonLabel}
          {ready && cloud !== 'gcp' ? <Icon variant="ArrowSquareOutIcon" size={14} /> : null}
        </Button>
      </div>
    </div>
  )
}

// --- Step 4 (all paths): how the install gets built ---------------------------
//
// Provisioning is slow, so this step explains the workflow instead of pretending
// to render it live. The live view is the next step: the install page.

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

interface IBuildStage {
  id: string
  icon: TIconVariant
  label: string
  text: string
  duration: string
}

const accountLabel = (path: TPath, cloud: TCloud) =>
  path === 'hosted' ? 'an AWS account Nuon runs' : `your ${CLOUD_CONNECT[cloud].accountNoun}`

// Strict linear order; the chain is the explanation for the duration. Copy is
// cloud-generic on purpose. Facts: the stack owns the network (the sandbox only
// tags its subnets), Nuon assumes four roles (provision, deprovision, maintenance,
// break-glass), and nuonco/aws-eks-sandbox provisions a cluster + node group,
// registry, storage and policy add-ons, DNS/ingress, namespaces, and RBAC.
// Durations from docs/get-started: network + machine + healthy runner ≈ 11 min;
// eks-simple end to end ≈ 35 min.
const buildStages = (path: TPath, cloud: TCloud, appName: string): IBuildStage[] => [
  {
    id: 'stack',
    icon: 'ShieldCheckIcon',
    label: 'Install stack',
    text:
      path === 'hosted'
        ? 'Network, runner machine, four roles — provision, deprovision, maintenance, break-glass. Nuon runs it in its own account; nothing for you to do.'
        : 'Network, runner machine, four roles — provision, deprovision, maintenance, break-glass. The one step you run yourself.',
    duration: 'about 10 min',
  },
  {
    id: 'runner',
    icon: 'CpuIcon',
    label: 'Runner',
    text: "Boots on the stack's machine and runs everything after this.",
    duration: 'about 1 min',
  },
  {
    id: 'sandbox',
    icon: 'StackIcon',
    label: 'Sandbox',
    text: 'Where components run: a cluster and nodes, registry, storage and policy add-ons, DNS and ingress, namespaces. Not always Kubernetes. The long one.',
    duration: 'about 15–20 min',
  },
  {
    id: 'components',
    icon: 'PackageIcon',
    label: 'Components',
    text: `${appName}'s Terraform, Helm, and images, deployed into the sandbox.`,
    duration: 'a few min',
  },
]

// The workflow, live. Stages before activeIndex are done, activeIndex is in
// progress, the rest show their typical duration — so the chain still answers
// "why is this slow" while it runs.
const BuildStages = ({ stages, activeIndex }: { stages: IBuildStage[]; activeIndex: number }) => (
  <ol className="grid gap-5 md:grid-cols-4 md:gap-0">
    {stages.map((stage, index) => {
      const state = index < activeIndex ? 'done' : index === activeIndex ? 'active' : 'next'
      return (
        <li key={stage.id} className="relative flex gap-4 md:flex-col md:items-center md:px-3 md:text-center">
          {index < stages.length - 1 ? (
            <span
              aria-hidden
              className={cn(
                'absolute left-1/2 top-6 hidden h-px w-full md:block',
                state === 'done' ? 'bg-green-500 dark:bg-green-400' : 'bg-neutral-200 dark:bg-neutral-600'
              )}
            />
          ) : null}
          <span
            className={cn(
              'relative z-10 flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-background',
              state === 'active'
                ? 'ring-2 ring-primary-500'
                : state === 'done'
                  ? 'ring-2 ring-green-500 dark:ring-green-400'
                  : 'ring-1 ring-neutral-200 dark:ring-neutral-600'
            )}
          >
            {state === 'done' ? (
              <Icon variant="CheckIcon" size={22} weight="bold" theme="success" />
            ) : (
              <Icon variant={stage.icon} size={22} theme={state === 'active' ? 'brand' : 'neutral'} />
            )}
          </span>
          <div className="flex flex-col gap-1.5 md:items-center">
            <Text variant="body" weight="strong">
              {index + 1}. {stage.label}
            </Text>
            {state === 'active' ? (
              <Badge size="sm" theme="brand">
                <Icon variant="Loading" size={12} /> In progress
              </Badge>
            ) : state === 'done' ? (
              <Badge size="sm" theme="success">
                Done
              </Badge>
            ) : (
              <Badge size="sm" theme="neutral">
                Up next · {stage.duration}
              </Badge>
            )}
            <Text variant="subtext" theme="neutral">
              {stage.text}
            </Text>
          </div>
        </li>
      )
    })}
  </ol>
)

// --- Step: the install workflow (the real multi-minute wait) --------------------
//
// The stack is creating; when it reports back the runner boots and the workflow
// takes over. In the product this is driven by install status; here a timer
// walks the chain. "See your install" ends the flow: the product opens the
// install's live workflow page.
const ProvisionStep = ({ sharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const path = readPath(sharedData)
  const cloud = readCloud(sharedData)
  const appName = path === 'own' ? readAppName(sharedData) : 'Kitchen Sink'
  const stages = buildStages(path, cloud, appName)
  const where = accountLabel(path, cloud)
  const [activeIndex, setActiveIndex] = useState(0)

  useEffect(() => {
    if (activeIndex >= stages.length - 1) return
    const timer = setTimeout(() => setActiveIndex((prev) => prev + 1), activeIndex === 0 ? 3000 : 2600)
    return () => clearTimeout(timer)
  }, [activeIndex, stages.length])

  return (
    <div className="flex flex-col gap-6">
      <Card className="!gap-0 !p-4 !flex-row items-center justify-between">
        <div className="flex items-center gap-3">
          {path === 'hosted' ? (
            <Icon variant="FlaskIcon" size={24} theme="brand" />
          ) : (
            <Icon variant={CLOUD_ICON[cloud]} size={24} />
          )}
          <div className="flex flex-col">
            <Text variant="base" weight="strong">
              {appName}
            </Text>
            <Text variant="body" theme="neutral">
              Building the install in {where} — {stages[activeIndex].label.toLowerCase()} in progress
            </Text>
          </div>
        </div>
        <Badge size="sm" theme={path === 'hosted' ? 'brand' : 'neutral'}>
          {path === 'hosted' ? 'Nuon-hosted account' : CLOUD_CONNECT[cloud].accountNoun}
        </Badge>
      </Card>

      <Card className="!gap-6">
        <div className="flex flex-col gap-1">
          <Text variant="h3" role="heading" level={3}>
            The Nuon install workflow
          </Text>
          <Text variant="body" theme="neutral">
            {path === 'hosted'
              ? 'Nuon created the stack in its own account. '
              : 'Your stack is creating. When it reports back, the runner boots and the workflow takes over. '}
            This creates all the cloud resources needed (network, VM, cluster) in {where}.
          </Text>
        </div>
        <BuildStages stages={stages} activeIndex={activeIndex} />
      </Card>

      <div className="flex flex-wrap items-center justify-between gap-3 rounded-md border px-4 py-3">
        <Text variant="body" theme="neutral">
          You can leave and come back — the install page shows live progress.
        </Text>
        <Text variant="subtext" theme="neutral" flex>
          Or watch from your terminal:
          <Badge size="sm" variant="code">
            nuon installs list
          </Badge>
        </Text>
      </div>

      <NextButton label="See your install" onClick={onAdvance} onBack={onGoBack} />
    </div>
  )
}

// --- Step 5 (all paths): the install page, live ---------------------------------

const InstallStep = ({ sharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const path = readPath(sharedData)
  const cloud = readCloud(sharedData)
  const rows = useMemo(() => provisionRows(path, cloud), [path, cloud])

  const [completed, setCompleted] = useState(0)
  const isDone = completed >= rows.length
  const activeRow = rows[completed]
  const appName = path === 'own' ? readAppName(sharedData) : 'kitchen-sink'

  useEffect(() => {
    if (isDone) return
    const timer = setTimeout(() => setCompleted((prev) => prev + 1), 950)
    return () => clearTimeout(timer)
  }, [completed, isDone])

  const liveHeading = isDone
    ? path === 'own'
      ? `${appName} is live`
      : path === 'hosted'
        ? 'Kitchen Sink is live in a Nuon-hosted account'
        : `Kitchen Sink is live in your ${CLOUD_CONNECT[cloud].accountNoun}`
    : `${activeRow.label} — ${activeRow.copy.active}`

  return (
    <div className="flex flex-col gap-6">
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
              {liveHeading}
            </Text>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {isDone ? (
            <Badge size="sm" theme="success">
              Active
            </Badge>
          ) : (
            <Badge size="sm" theme="brand">
              <Icon variant="Loading" size={12} /> Provisioning
            </Badge>
          )}
          {path === 'hosted' ? (
            <Badge size="sm" theme="brand">
              Nuon-hosted account
            </Badge>
          ) : (
            <Badge size="sm" theme="neutral">
              {path === 'own' ? 'Your cloud account' : CLOUD_CONNECT[cloud].accountNoun}
            </Badge>
          )}
        </div>
      </Card>

      <div className="flex items-center justify-between">
        <Text variant="base" weight="strong">
          Resources
        </Text>
        <Text variant="body" theme="neutral">
          {isDone ? 'All resources provisioned' : `${completed} of ${rows.length} ready`}
        </Text>
      </div>

      <Card className="!gap-0 !p-0 overflow-hidden">
        {rows.map((row, index) => {
          const rowDone = index < completed
          const rowActive = index === completed
          const status = rowDone ? row.copy.done : rowActive ? row.copy.active : row.copy.pending

          return (
            <div key={row.id} className={cn('flex items-center gap-3 px-5 py-3', index > 0 && 'border-t')}>
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

      <Text variant="subtext" theme="neutral">
        {isDone
          ? 'Everything is up. This is the page your customer sees for their install.'
          : 'This page updates on its own. You can leave and come back.'}
      </Text>
      <NextButton label="Go to dashboard" onClick={onAdvance} onBack={onGoBack} />
    </div>
  )
}


// --- Parked: the "is live" summary page ------------------------------------------
// Matt likes this design but wants it out of the first critical path. It is not
// in any flow; the ParkedDone story keeps it reviewable.

const DoneStep = ({ sharedData, onAdvance }: IWizardStepComponentProps) => {
  const path = readPath(sharedData)
  const cloud = readCloud(sharedData)
  const appName = readAppName(sharedData)

  const heading =
    path === 'own'
      ? `${appName} is live`
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

const PARKED_DONE_STEP: IWizardStepDef = {
  id: 'parked-done',
  title: "You're all set",
  navLabel: 'Done',
  hideTitle: true,
  component: DoneStep,
}

// --- Step definitions ---------------------------------------------------------

// The step renders its own, larger title (no subhead) instead of the wizard's default h2.
const FORK_STEP: IWizardStepDef = {
  id: 'fork',
  title: 'Create your first app template',
  navLabel: 'Start',
  hideTitle: true,
  component: ForkStep,
}

const TEMPLATE_STEP: IWizardStepDef = {
  id: 'own-template',
  title: 'Create your app template',
  navLabel: 'Template',
  description: 'Your app template is what Nuon uses to deploy your product.',
  component: TemplateStep,
}

const DEPLOY_STEP: IWizardStepDef = {
  id: 'deploy',
  title: 'Set up your first install',
  navLabel: 'Deploy',
  description: 'Nothing has touched your account yet. Choose where the install goes.',
  component: DeployStep,
}

const INSTALL_STACK_STEP: IWizardStepDef = {
  id: 'install-stack',
  title: 'Create the install stack',
  navLabel: 'Stack',
  description:
    'Nuon is generating your stack link — about 30 seconds. While it does, here is how a customer would create the stack.',
  component: StackStep,
}

const PROVISION_STEP: Record<TPath, IWizardStepDef> = {
  example: {
    id: 'example-provision',
    title: 'Your install is being created',
    navLabel: 'Provision',
    description: 'The Nuon install workflow, under the hood.',
    component: ProvisionStep,
  },
  hosted: {
    id: 'hosted-provision',
    title: 'Your install is being created',
    navLabel: 'Provision',
    description: 'The Nuon install workflow, under the hood — in an account Nuon runs.',
    component: ProvisionStep,
  },
  own: {
    id: 'own-provision',
    title: 'Your install is being created',
    navLabel: 'Provision',
    description: 'The Nuon install workflow, under the hood.',
    component: ProvisionStep,
  },
}

// Parked: the live install page. "See your install" opens the install's workflow
// page in the product, so this is no longer a step in the flow.
const PARKED_INSTALL_STEP: IWizardStepDef = {
  id: 'parked-install',
  title: 'Your install',
  navLabel: 'Install',
  description: 'Live from the runner. This is the page your customer sees for their install.',
  component: InstallStep,
}

const buildForkFlow = (path: TPath, cloud: TCloud): IWizardStepDef[] => {
  if (path === 'hosted') return [FORK_STEP, PROVISION_STEP.hosted]
  if (path === 'own') return [FORK_STEP, TEMPLATE_STEP, DEPLOY_STEP, INSTALL_STACK_STEP, PROVISION_STEP.own]
  return [
    FORK_STEP,
    { ...DEPLOY_STEP, id: `deploy-${cloud}` },
    { ...INSTALL_STACK_STEP, id: `install-stack-${cloud}` },
    { ...PROVISION_STEP.example, id: `example-provision-${cloud}` },
  ]
}

// --- Copy editor (review tooling — strip before this leaves the playground) ---
//
// Wraps the wizard so any visible text can be retyped directly at :61000, then
// handed back as a JSON list of { from, to } pairs. Headings, body copy, links,
// and button labels are all editable (verified in Chromium). While editing is on,
// clicks on buttons and links are swallowed so placing a caret does not navigate.

interface ICopyEdit {
  from: string
  to: string
  tag: string
}

interface ICopyOriginal {
  own: string
  full: string
}

const COPY_EDITOR_SKIP = new Set(['SCRIPT', 'STYLE', 'SVG', 'PATH', 'INPUT', 'TEXTAREA', 'SELECT', 'OPTION'])

const normalizeCopy = (value: string) => value.replace(/\s+/g, ' ').trim()

const ownCopy = (el: Element) =>
  normalizeCopy(
    Array.from(el.childNodes)
      .filter((node) => node.nodeType === Node.TEXT_NODE)
      .map((node) => node.nodeValue ?? '')
      .join('')
  )

const elementOf = (node: Node | null | undefined): Element | null => {
  if (!node) return null
  return node.nodeType === Node.TEXT_NODE ? node.parentElement : (node as Element)
}

const CopyEditor = ({ children, tools }: { children: ReactNode; tools?: ReactNode }) => {
  const wrapRef = useRef<HTMLDivElement>(null)
  // React reuses DOM nodes across steps (the step heading is one <h2> whose text
  // changes), so an element's "original" is only valid until React rewrites it.
  // Live edits are keyed by element; once React repurposes or detaches the
  // element, the edit is moved to `committed` and the original is refreshed.
  const tracked = useRef(new Set<Element>())
  const originals = useRef(new WeakMap<Element, ICopyOriginal>())
  const live = useRef(new Map<Element, ICopyEdit & { toOwn: string }>())
  const committed = useRef<ICopyEdit[]>([])
  const [editing, setEditing] = useState(false)
  const [count, setCount] = useState(0)
  const [copied, setCopied] = useState(false)

  const snapshot = (el: Element) => {
    originals.current.set(el, { own: ownCopy(el), full: normalizeCopy(el.textContent ?? '') })
    tracked.current.add(el)
  }

  const remember = (el: Element) => {
    if (COPY_EDITOR_SKIP.has(el.tagName) || originals.current.has(el)) return
    if (ownCopy(el)) snapshot(el)
  }

  // If the element's text no longer matches what we last saw (our edit, or the
  // original), React rewrote it: finalize any live edit and start fresh.
  const resync = (el: Element, root: HTMLElement) => {
    const original = originals.current.get(el)
    if (!original) return
    const edit = live.current.get(el)
    const detached = !root.contains(el)
    const rewritten = ownCopy(el) !== (edit?.toOwn ?? original.own)
    if (!detached && !rewritten) return
    if (edit) {
      const { toOwn: _toOwn, ...done } = edit
      committed.current.push(done)
      live.current.delete(el)
    }
    if (detached) {
      tracked.current.delete(el)
      originals.current.delete(el)
    } else {
      snapshot(el)
    }
  }

  const resyncAll = (root: HTMLElement) => {
    Array.from(tracked.current).forEach((el) => resync(el, root))
    setCount(committed.current.length + live.current.size)
  }

  useEffect(() => {
    const root = wrapRef.current?.firstElementChild as HTMLElement | null
    if (!root || !editing) return

    const lineage = (node: Node | null | undefined) => {
      const out: Element[] = []
      let el = elementOf(node)
      while (el && el !== root) {
        out.push(el)
        el = el.parentElement
      }
      return out
    }

    // Placing a caret in a button or link must not trigger it.
    const swallowClicks = (event: MouseEvent) => {
      if ((event.target as Element).closest('button, a, summary, [role="button"]')) {
        event.preventDefault()
        event.stopPropagation()
      }
    }

    // Keep edits inside one text block: no new paragraphs, no formatting, no cross-element deletes.
    const guardInput = (event: InputEvent) => {
      const selection = document.getSelection()
      if (
        event.inputType.startsWith('format') ||
        event.inputType === 'insertParagraph' ||
        event.inputType === 'insertLineBreak'
      ) {
        event.preventDefault()
        return
      }
      if (
        selection &&
        !selection.isCollapsed &&
        elementOf(selection.anchorNode) !== elementOf(selection.focusNode)
      ) {
        event.preventDefault()
        return
      }
      lineage(selection?.anchorNode).forEach((el) => {
        remember(el)
        resync(el, root)
      })
    }

    const pastePlain = (event: ClipboardEvent) => {
      event.preventDefault()
      const text = event.clipboardData?.getData('text/plain') ?? ''
      document.execCommand('insertText', false, normalizeCopy(text))
    }

    const record = () => {
      const selection = document.getSelection()
      const touched = new Set<Element>([
        ...lineage(selection?.anchorNode),
        ...lineage(selection?.focusNode),
      ])
      touched.forEach((el) => {
        const original = originals.current.get(el)
        if (!original) return
        const toOwn = ownCopy(el)
        if (toOwn === original.own) {
          live.current.delete(el)
        } else {
          live.current.set(el, {
            from: original.full,
            to: normalizeCopy(el.textContent ?? ''),
            tag: el.tagName.toLowerCase(),
            toOwn,
          })
        }
      })
      setCount(committed.current.length + live.current.size)
    }

    resyncAll(root)
    root.querySelectorAll('*').forEach(remember)
    root.contentEditable = 'true'
    root.spellcheck = false
    root.style.cursor = 'text'
    root.addEventListener('click', swallowClicks, true)
    root.addEventListener('beforeinput', guardInput)
    root.addEventListener('paste', pastePlain)
    root.addEventListener('input', record)

    return () => {
      root.removeEventListener('click', swallowClicks, true)
      root.removeEventListener('beforeinput', guardInput)
      root.removeEventListener('paste', pastePlain)
      root.removeEventListener('input', record)
      root.removeAttribute('contenteditable')
      root.removeAttribute('spellcheck')
      root.style.cursor = ''
    }
  }, [editing])

  const collect = () => {
    const root = wrapRef.current?.firstElementChild as HTMLElement | null
    if (root) resyncAll(root)
    const current = Array.from(live.current.values()).map(({ toOwn: _toOwn, ...edit }) => edit)
    return JSON.stringify([...committed.current, ...current], null, 2)
  }

  const copyEdits = () => {
    navigator.clipboard?.writeText(collect()).catch(() => {})
    setCopied(true)
    window.setTimeout(() => setCopied(false), 2000)
  }

  const downloadEdits = () => {
    const url = URL.createObjectURL(new Blob([collect()], { type: 'application/json' }))
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = 'nuon-copy-edits.json'
    anchor.click()
    window.setTimeout(() => URL.revokeObjectURL(url), 1000)
  }

  const clearEdits = () => {
    live.current.clear()
    committed.current = []
    tracked.current.clear()
    originals.current = new WeakMap()
    setCount(0)
  }

  const nothingYet = count === 0
  const editsLabel = `${count} ${count === 1 ? 'edit' : 'edits'}`

  return (
    <>
      <div ref={wrapRef} className="contents">
        {children}
      </div>
      {/* Bottom-center: every step ends in a Back/Next row whose middle is empty, so the
          panel never covers a button. Bottom-right sat on top of right-aligned primaries. */}
      <Card className="fixed bottom-4 left-1/2 z-50 max-w-[calc(100vw-2rem)] -translate-x-1/2 !flex-row flex-wrap items-center justify-center !gap-3 !p-3 bg-background">
        {tools ? (
          <>
            {tools}
            <span className="h-5 w-px bg-neutral-200 dark:bg-neutral-600" aria-hidden />
          </>
        ) : null}
        <Toggle checked={editing} onChange={setEditing} label={editing ? 'Editing copy' : 'Edit copy'} />
        <Button
          variant="secondary"
          size="sm"
          onClick={copyEdits}
          disabled={nothingYet}
          tooltipProps={nothingYet ? { tipContent: 'Cannot copy edits — no text changed yet' } : undefined}
        >
          <Icon variant={copied ? 'CheckIcon' : 'CopyIcon'} weight="bold" />
          {copied ? 'Copied' : `Copy ${editsLabel}`}
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={downloadEdits}
          disabled={nothingYet}
          tooltipProps={nothingYet ? { tipContent: 'Cannot download edits — no text changed yet' } : undefined}
        >
          <Icon variant="DownloadSimpleIcon" weight="bold" />
          Download
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={clearEdits}
          disabled={nothingYet}
          tooltipProps={nothingYet ? { tipContent: 'Cannot clear edits — nothing recorded yet' } : undefined}
        >
          Clear
        </Button>
      </Card>
    </>
  )
}

// --- Harness ------------------------------------------------------------------

const TEMPLATE_LAYOUTS: { value: TTemplateLayout; label: string }[] = [
  { value: 'top', label: 'Top' },
  { value: 'beside', label: 'Beside' },
]

const BranchingPlayground = ({
  initialPath = 'example',
  initialCloud = 'aws',
  initialStepIndex = 0,
  skipIntro = false,
  expandOwnApp = false,
}: {
  initialPath?: TPath
  initialCloud?: TCloud
  initialStepIndex?: number
  skipIntro?: boolean
  expandOwnApp?: boolean
}) => {
  const [runId, setRunId] = useState(0)
  // Deep-linked stories skip the intro; the default story starts on it.
  const [started, setStarted] = useState(initialStepIndex > 0 || skipIntro)
  const [startIndex, setStartIndex] = useState(initialStepIndex)
  const [finished, setFinished] = useState(false)
  const [choice, setChoice] = useState<IForkChoice>({ path: initialPath, cloud: initialCloud })
  // Review-only: where the template step puts the stubbed files.
  const [templateLayout, setTemplateLayout] = useState<TTemplateLayout>('top')
  // Review-only: stands in for a git push to the tracked branch.
  const [pushes, setPushes] = useState(0)

  const steps = useMemo(
    () => buildForkFlow(choice.path, choice.cloud ?? initialCloud),
    [choice, initialCloud]
  )

  const reset = (toIntro: boolean) => {
    setFinished(false)
    setStarted(!toIntro)
    setStartIndex(toIntro ? 0 : initialStepIndex)
    setChoice({ path: initialPath, cloud: initialCloud })
    setPushes(0)
    setRunId((prev) => prev + 1)
  }

  const fork: IForkActions = { choose: setChoice, backToIntro: () => reset(true), templateLayout, pushTick: pushes }

  if (finished) {
    return <FlowComplete onRestart={() => reset(initialStepIndex > 0 || skipIntro)} />
  }

  return (
    <CopyEditor
      tools={
        <div className="flex items-center gap-2">
          <Text variant="subtext" theme="neutral">
            Files
          </Text>
          <ToggleButton<TTemplateLayout> options={TEMPLATE_LAYOUTS} value={templateLayout} onChange={setTemplateLayout} size="sm" />
          <Button variant="ghost" size="sm" onClick={() => setPushes((prev) => prev + 1)}>
            <Icon variant="GitBranchIcon" size={14} /> Simulate push
          </Button>
        </div>
      }
    >
      {started ? (
        <ForkContext.Provider value={fork}>
          <OnboardingWizardProvider
            key={runId}
            steps={steps}
            initialStepIndex={startIndex}
            initialSharedData={{ path: initialPath, cloud: initialCloud, expandOwn: expandOwnApp }}
            onComplete={() => setFinished(true)}
          >
            <OnboardingWizardLayout skipHref={null} />
          </OnboardingWizardProvider>
        </ForkContext.Provider>
      ) : (
        <IntroScreen onStart={() => setStarted(true)} />
      )}
    </CopyEditor>
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

export const ForkOwnApp = () => <BranchingPlayground initialPath="own" skipIntro expandOwnApp />
ForkOwnApp.meta = { fullBleed: true }

// Not part of any flow. Kept so the design is still reviewable.
export const ParkedDone = () => (
  <OnboardingWizardProvider
    steps={[PARKED_DONE_STEP]}
    initialSharedData={{ path: 'own', cloud: 'aws' }}
    onComplete={() => {}}
  >
    <OnboardingWizardLayout skipHref={null} />
  </OnboardingWizardProvider>
)
ParkedDone.meta = { fullBleed: true }

export const ParkedInstall = () => (
  <OnboardingWizardProvider
    steps={[PARKED_INSTALL_STEP]}
    initialSharedData={{ path: 'own', cloud: 'aws' }}
    onComplete={() => {}}
  >
    <OnboardingWizardLayout skipHref={null} />
  </OnboardingWizardProvider>
)
ParkedInstall.meta = { fullBleed: true }
