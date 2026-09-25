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
  secondary,
}: {
  label?: string
  disabled?: boolean
  disabledReason?: string
  onClick?: () => void
  onBack?: () => void
  showNext?: boolean
  // Rendered beside the primary (e.g. a cost line while a push is still pending).
  secondary?: ReactNode
}) => (
  <div className={cn('flex gap-3', onBack ? 'justify-between' : 'justify-end')}>
    {onBack ? (
      <Button variant="secondary" onClick={onBack}>
        <Icon variant="CaretLeftIcon" weight="bold" /> Back
      </Button>
    ) : null}
    <div className="flex items-center gap-3">
      {secondary}
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
// Fork flow: post-login screen that splits into two paths.
//   example → deploy Kitchen Sink into the user's own cloud (stack link, pre-filled inputs)
//   own     → connect GitHub, set the app up from the terminal or an MCP agent
// The steps array is swapped when the fork picks a path; the provider reads
// `steps` from props on every render so the stepper follows the chosen path.
// An intro page sits before the stepper. The fork leads with the user's own app
// as the single primary; the example app is the secondary path.
// ---------------------------------------------------------------------------

type TCloud = 'aws' | 'gcp' | 'azure'
type TPath = 'example' | 'own'

interface IForkChoice {
  path: TPath
  cloud?: TCloud
}

interface IForkActions {
  choose: (choice: IForkChoice) => void
  backToIntro: () => void
  // Review-only: bumps each time "Simulate push" is pressed; the product's trigger is the push itself.
  pushTick: number
}

const ForkContext = createContext<IForkActions>({
  choose: () => {},
  backToIntro: () => {},
  pushTick: 0,
})
const useForkChoice = () => useContext(ForkContext)

// Clouds the example app (Kitchen Sink) can be deployed to from the fork. The
// own-app path offers all three.
const EXAMPLE_CLOUDS: TCloud[] = ['aws', 'gcp']

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
    // What exists once generation finishes. AWS gets a console link; GCP and Azure (at its
    // default scope) get a template and commands, so the status line must not say "link".
    artifactNoun: string
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
    artifactNoun: 'CloudFormation stack link',
    generating: 'Generating your CloudFormation stack link...',
    launch: 'Open the CloudFormation stack',
    opening: 'Opening the AWS console...',
    helper:
      'Opens a pre-filled CloudFormation stack in your AWS console. Create it there, then come back. This page updates on its own.',
    waitingHint: 'Create the CloudFormation stack in the AWS console tab, then come back.',
  },
  gcp: {
    accountNoun: 'GCP project',
    stackLabel: 'Terraform stack',
    artifactNoun: 'Terraform stack',
    generating: 'Generating your Terraform stack...',
    launch: 'Get the Terraform stack',
    opening: 'Preparing the Terraform stack...',
    helper:
      'Nuon generates a Terraform stack for your test GCP project. Apply it from your terminal, then come back. This page updates on its own.',
    waitingHint: 'Apply the Terraform stack from your terminal, then come back.',
  },
  azure: {
    accountNoun: 'Azure subscription',
    stackLabel: 'Bicep stack',
    artifactNoun: 'Bicep template',
    generating: 'Generating your Bicep template...',
    launch: 'Get the Azure commands',
    opening: 'Preparing the commands...',
    helper:
      'Nuon generates the Bicep template and the az commands that deploy it. Create the resource group and Key Vault, run the commands, then come back. This page updates on its own.',
    waitingHint: 'Run the az commands from your terminal, then come back.',
  },
}

const CLOUD_REGIONS: Record<TCloud, { label: string; options: string[] }> = {
  aws: {
    label: 'AWS region',
    options: ['us-east-1', 'us-west-2', 'eu-west-1', 'ap-southeast-1'],
  },
  gcp: {
    label: 'GCP region',
    options: ['us-central1', 'us-east1', 'europe-west1', 'asia-southeast1'],
  },
  azure: {
    label: 'Azure location',
    options: ['eastus', 'westus2', 'westeurope', 'southeastasia'],
  },
}

const regionOptions = (cloud: TCloud) =>
  CLOUD_REGIONS[cloud].options.map((value) => ({ value, label: value }))

// The sandbox each cloud lands on, and the repo that provisions it.
const CLOUD_SANDBOX: Record<TCloud, string> = {
  aws: 'nuonco/aws-eks-sandbox',
  gcp: 'nuonco/gcp-gke-sandbox',
  azure: 'nuonco/azure-aks-sandbox',
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

const IntroDiagram = () => (
  <div className="flex flex-col gap-3">
    <div className="flex flex-col gap-3 rounded-lg border bg-background p-4 shadow-sm">
      <div className="flex items-center gap-2">
        <Icon variant="GitBranchIcon" size={18} theme="neutral" />
        <Text variant="base" weight="strong">
          Your app template
        </Text>
      </div>
      <MiniArch />
    </div>

    <div className="flex items-center justify-center gap-3 py-1">
      <NuonMark className="h-7 w-auto text-neutral-900 dark:text-white" />
      <Icon variant="ArrowDownIcon" size={24} weight="bold" theme="neutral" />
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
      </div>
    </div>
  </div>
)

const IntroScreen = ({ onStart }: { onStart: () => void }) => (
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
          <div>
            <Button variant="primary" size="lg" onClick={onStart}>
              Create your first app template <Icon variant="CaretRightIcon" weight="bold" />
            </Button>
          </div>
        </div>
        <IntroDiagram />
      </div>
    </div>
  </div>
)

// --- Step 1: the fork --------------------------------------------------------
//
// "Start with your app" does not advance the wizard. It expands the setup in
// place: connect GitHub, install the CLI, then create the app. On the Template
// step the agent path is the whole card; manual setup is a cautioned option
// below it, beside a way to have Nuon's team write the config.
//
// Order, verified against docs/guides/agents: the MCP server is `nuon agents
// mcp`, a CLI subcommand, so the CLI must be installed and logged in first. The
// nuon-loop paste drives the CLI directly and does not need the MCP server; MCP
// is offered as an optional extra. GitHub is only required for `connected_repo`
// components (private repos); production onboarding v2 has no GitHub step.

// Where "Contact us" lands. In the product, AuthLayout loads the Pylon chat widget
// for signed-in users (lib/pylon-chat), so the button opens it with a message
// started. The playground has no widget, so it falls back to the demo form.
const DEMO_REQUEST = 'https://nuon.co/demo-request'
const CONTACT_MESSAGE = 'I would like help writing the app config for my first install.'
const contactUs = () => {
  if (typeof window.Pylon === 'function') {
    window.Pylon('showNewMessage', CONTACT_MESSAGE)
    return
  }
  window.open(DEMO_REQUEST, '_blank', 'noopener,noreferrer')
}

const DOCS_MCP = 'https://docs.nuon.co/guides/agents/mcp-walkthrough'
// Placeholder until the marketing site hosts the prompt as its own text file.
const PROMPT_TXT_URL = 'https://nuon.co/llms.txt'
const DOCS_RUNNERS = 'https://docs.nuon.co/concepts/runners'
const DOCS_SANDBOXES = 'https://docs.nuon.co/concepts/sandboxes'
const CLI_SETUP = 'brew install nuonco/tap/nuon\nnuon login'
const MCP_ADD_CLAUDE = 'claude mcp add --transport stdio nuon -- nuon agents mcp --allow-writes'
const DOCS_CONFIG_FILES = 'https://docs.nuon.co/configuration-files'
const AWS_QUICK_CREATE_DOCS =
  'https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cfn-console-create-stacks-quick-create-links.html'
const GCP_INFRA_MANAGER_DOCS = 'https://cloud.google.com/infrastructure-manager/docs'
const GIT_PUSH = (app: string) => `git add ${app}\ngit commit -m "Add Nuon app config"\ngit push origin main`
const VSCODE_EXTENSION = 'https://marketplace.visualstudio.com/items?itemName=Nuon.nuon-lsp'
const LSP_NEOVIM_SETUP = 'https://github.com/nuonco/nuon/blob/main/bins/lsp/README.md#neovim'
const DOCS_LSP = 'https://docs.nuon.co/configuration-files#language-server-protocol-lsp'
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
  variant = 'secondary',
}: {
  text: string
  label: string
  size?: 'lg' | 'md' | 'sm'
  variant?: 'primary' | 'secondary'
}) => {
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!copied) return
    const timer = setTimeout(() => setCopied(false), 2000)
    return () => clearTimeout(timer)
  }, [copied])

  return (
    <Button
      variant={variant}
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

const OWN_APP_STEPS: { icon: TIconVariant; title: string }[] = [
  { icon: 'GitHub', title: 'Connect GitHub' },
  { icon: 'RobotIcon', title: 'Connect your app' },
  { icon: 'CloudIcon', title: 'Create the first install' },
]

// The step's live moment: Nuon watching the tracked branch. In the prototype the
// review panel's "Simulate push" stands in for the push.
const PushListener = ({ detected, cloud }: { detected: boolean; cloud: TCloud }) => (
  <div
    className={cn(
      'flex items-start justify-between gap-3 rounded-md bg-background p-4 ring-1 transition-shadow',
      detected ? 'ring-green-500 dark:ring-green-400' : 'ring-neutral-200 dark:ring-neutral-700'
    )}
  >
    <div className="flex min-w-0 flex-1 items-start gap-3">
      {detected ? (
        <Icon variant="CheckCircleIcon" size={20} weight="fill" theme="success" />
      ) : (
        <Icon variant="Loading" size={20} />
      )}
      <div className="flex min-w-0 flex-col gap-0.5">
        <Text variant="body" weight="strong">
          {detected ? 'Synced from main' : 'The prompt ends with pushing your app config'}
        </Text>
        {detected ? (
          <Text variant="subtext" theme="neutral" flex className="flex-wrap">
            Commit
            <Badge size="sm" variant="code">
              a1b2c3d
            </Badge>
            updated the default app branch. Building your components now.
          </Text>
        ) : (
          <Text variant="subtext" theme="neutral">
            Continue now and the next steps create the install and provision the{' '}
            <Link href={DOCS_RUNNERS} isExternal textVariant="subtext" className="!inline-flex align-baseline">
              runner
            </Link>{' '}
            and{' '}
            <Link href={DOCS_SANDBOXES} isExternal textVariant="subtext" className="!inline-flex align-baseline">
              Nuon sandbox
            </Link>{' '}
            in your {CLOUD_LABEL[cloud]} test account. Your app deploys when the push lands. Or wait for the push and
            watch it all deploy in one workflow.
          </Text>
        )}
      </div>
    </div>
    <Badge size="sm" theme={detected ? 'success' : 'brand'} className="mt-0.5 shrink-0">
      {detected ? 'Synced' : 'Watching'}
    </Badge>
  </div>
)

interface IAppFileStub {
  name: string
  purpose: string
  badge: string
  required: boolean
  snippet: (app: string, cloud: TCloud) => string
}

// Policy shapes per cloud, from nuonco/example-app-configs: kitchen-sink (AWS managed
// policy), gke-simple (gcp_predefined_role), aks-simple (azure_built_in_roles).
const ROLE_POLICY: Record<TCloud, string[]> = {
  aws: ['managed_policy_name = "AdministratorAccess"'],
  gcp: ['name                = "owner"', 'gcp_predefined_role = "roles/owner"'],
  azure: ['name                 = "contributor"', 'azure_built_in_roles = ["Contributor"]'],
}
const PERMISSIONS_STUB = (cloud: TCloud) =>
  (
    [
      ['provision', 'Provision the sandbox and components.'],
      ['maintenance', 'Operate and update components.'],
      ['deprovision', 'Tear the install down.'],
    ] as const
  )
    .map(([role, description]) =>
      [
        `[${role}_role]`,
        `name        = "{{.nuon.install.id}}-${role}"`,
        `description = "${description}"`,
        ...(cloud === 'aws' ? [] : [`cloud_platform = "${cloud}"`]),
        `[[${role}_role.policies]]`,
        ...ROLE_POLICY[cloud],
      ].join('\n')
    )
    .join('\n\n')

const APP_FILE_STUBS: IAppFileStub[] = [
  {
    name: 'metadata.toml',
    purpose: 'Names the app and pins the config version.',
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
    snippet: (_app, cloud) => `# aws, azure, or gcp\nrunner_type = "${cloud}"`,
  },
  {
    name: 'sandbox.toml',
    purpose: 'The base infrastructure the app lands on: the Nuon sandbox for your test cloud.',
    badge: 'Required',
    required: true,
    snippet: (_app, cloud) =>
      `terraform_version = "1.11.3"\n\n[public_repo]\nrepo      = "${CLOUD_SANDBOX[cloud]}"\ndirectory = "."\nbranch    = "main"`,
  },
  {
    name: 'permissions.toml',
    purpose: 'The roles Nuon assumes in the customer account: provision, maintenance and deprovision.',
    badge: 'Required',
    required: true,
    snippet: (_app, cloud) => PERMISSIONS_STUB(cloud),
  },
  {
    name: 'branch.toml',
    purpose: 'Tracks the repo you connected. Every push to main syncs the default app branch.',
    badge: 'Created for you',
    required: false,
    snippet: (app) =>
      `name = "default"\n\n[connected_repo]\nrepo      = "jane-doe/${app}"\ndirectory = "."\nbranch    = "main"`,
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

// One expandable row per file.
const FileStubRows = ({ appName, cloud }: { appName: string; cloud: TCloud }) => {
  const shown = useMountedReveal()
  const [open, setOpen] = useState<string[]>([])
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
                  {file.snippet(appName, cloud)}
                </CodeBlock>
              </div>
            ) : null}
          </li>
        )
      })}
    </ul>
  )
}


const EXAMPLE_APP_FACTS = [
  'Helm chart: API, UI, and worker pods',
  'Pulumi S3 bucket and CI-built images',
  'Actions, policies, runbooks, app branches',
]
const ExampleAppDrawer = () => {
  const [open, setOpen] = useState(false)
  return (
    <div className="flex flex-col rounded-md border border-dashed">
      <button
        type="button"
        aria-expanded={open}
        aria-controls="example-app-drawer"
        onClick={() => setOpen((prev) => !prev)}
        className="flex w-full items-center justify-between gap-3 px-4 py-2.5 text-left hover:bg-neutral-50 dark:hover:bg-neutral-900"
      >
        <span className="flex items-center gap-2">
          <Icon variant="GithubLogoIcon" size={16} theme="neutral" />
          <Text variant="subtext" weight="strong">
            See the example app repo
          </Text>
          <Badge size="sm" variant="code">nuonco/kitchen-sink</Badge>
        </span>
        <span className={cn('flex transition-transform duration-300', open && 'rotate-180')} aria-hidden>
          <Icon variant="CaretDownIcon" size={14} weight="bold" theme="neutral" />
        </span>
      </button>
      <div
        id="example-app-drawer"
        className={cn(
          'grid transition-[grid-template-rows,opacity,visibility] duration-300 ease-out',
          open ? 'visible grid-rows-[1fr] opacity-100' : 'invisible grid-rows-[0fr] opacity-0'
        )}
        aria-hidden={!open}
      >
        <div className="overflow-hidden">
          <div className="flex flex-col gap-3 border-t border-dashed px-4 py-3 md:flex-row md:items-start md:justify-between">
            <div className="flex flex-col gap-1.5">
              <Text variant="body" weight="strong">
                Kitchen Sink
              </Text>
              <ul className="flex flex-col gap-1">
                {EXAMPLE_APP_FACTS.map((fact) => (
                  <li key={fact} className="flex items-start gap-2">
                    <Icon variant="CheckCircleIcon" size={14} weight="fill" theme="success" className="mt-0.5 shrink-0" />
                    <Text variant="subtext" theme="neutral">
                      {fact}
                    </Text>
                  </li>
                ))}
              </ul>
            </div>
            <Button variant="secondary" size="sm" href={KITCHEN_SINK_REPO} target="_blank" rel="noreferrer">
              <Icon variant="GithubLogoIcon" size={14} /> View on GitHub
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

const ExampleEscapeHatch = ({ onExit }: { onExit: () => void }) => (
  <div className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-dashed px-4 py-3">
    <div className="flex items-center gap-2">
      <Icon variant="TireIcon" size={16} theme="neutral" />
      <Text variant="subtext" theme="neutral">
        Want to see an install work before touching your repo?
      </Text>
    </div>
    <Button variant="ghost" size="sm" onClick={onExit}>
      Use the example app <Icon variant="ArrowRightIcon" size={14} />
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
  cloud,
  onCloud,
}: {
  heading: ReactNode
  appName: string
  onAppName: (name: string) => void
  githubDone: boolean
  onGithubDone: () => void
  showErrors: boolean
  cloud?: TCloud
  onCloud: (cloud: TCloud) => void
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
      · 3 repos
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
            error={showErrors && !named}
            errorMessage="Name your app template to continue."
            autoComplete="off"
            spellCheck={false}
          />
        </div>
        {/* Asked here, before the template step, so the stubbed runner, sandbox and
            permissions match the cloud the install will use. */}
        <TestCloudPicker value={cloud} onChange={onCloud} error={showErrors && !cloud} />
      </Card>
    </>
  )
}

// --- Step 1b: the template (own path only) ------------------------------------
//
// The app exists and its config is stubbed. This step shows the stubs and the
// two ways to fill them in.
// The optional reading, as one line of fine print under the step. NN/g's progressive
// disclosure: show the primary task, disclose the rest only when asked, with labels
// that say what opens. Grey, dotted underline, no border: nothing here competes with
// "Copy prompt".
type TFootnote = 'mcp' | 'deps' | 'manual'
const FOOTNOTE_LINK =
  'cursor-pointer text-cool-grey-500 underline decoration-dotted underline-offset-2 hover:text-foreground dark:text-cool-grey-400'
const Footnotes = ({
  appName,
  repo,
  cloud,
  onExampleExit,
}: {
  appName: string
  repo: string
  cloud: TCloud
  onExampleExit: () => void
}) => {
  const [open, setOpen] = useState<TFootnote | null>(null)
  const toggle = (key: TFootnote) => setOpen((prev) => (prev === key ? null : key))
  return (
    <div className="flex flex-col gap-3">
      <Text variant="subtext" theme="neutral" flex className="flex-wrap gap-x-2">
        <span>Optional:</span>
        <button
          type="button"
          aria-expanded={open === 'mcp'}
          aria-controls="footnote-mcp"
          onClick={() => toggle('mcp')}
          className={FOOTNOTE_LINK}
        >
          MCP setup
        </button>
        <span aria-hidden>·</span>
        <button
          type="button"
          aria-expanded={open === 'deps'}
          aria-controls="footnote-deps"
          onClick={() => toggle('deps')}
          className={FOOTNOTE_LINK}
        >
          Dependencies
        </button>
        <span aria-hidden>·</span>
        <button
          type="button"
          aria-expanded={open === 'manual'}
          aria-controls="footnote-manual"
          onClick={() => toggle('manual')}
          className={FOOTNOTE_LINK}
        >
          Manual steps
        </button>
        <span aria-hidden>·</span>
        <button type="button" onClick={contactUs} className={FOOTNOTE_LINK}>
          Get help
        </button>
        <span aria-hidden>·</span>
        <button type="button" onClick={onExampleExit} className={FOOTNOTE_LINK}>
          Use the example app instead
        </button>
      </Text>
      {open === 'manual' ? (
        <div id="footnote-manual" className="rounded-md border bg-background p-4">
          <ManualSetup appName={appName} repo={repo} cloud={cloud} />
        </div>
      ) : null}
      {open === 'mcp' ? (
        <div id="footnote-mcp" className="flex flex-col gap-1.5 rounded-md border bg-background p-4">
          <Text variant="subtext" weight="strong">
            Give your agent the Nuon Model Context Protocol (MCP) server
          </Text>
          <Text variant="subtext" theme="neutral">
            Live access to your org while it works: apps, builds, installs and logs. In Claude Code:
          </Text>
          <CodeBlock language="bash" showCopy wrapLongLines className="!pr-14">
            {MCP_ADD_CLAUDE}
          </CodeBlock>
          <Link href={DOCS_MCP} isExternal textVariant="subtext">
            docs.nuon.co/guides/agents/mcp-walkthrough
          </Link>
        </div>
      ) : null}
      {open === 'deps' ? (
        <div id="footnote-deps" className="flex flex-col gap-1.5 rounded-md border bg-background p-4">
          <Text variant="subtext" weight="strong">
            Dependencies
          </Text>
          <Text variant="subtext" theme="neutral">
            Databases like Postgres run as components in the customer&apos;s account. Third-party services like
            Clerk or SendGrid stay external; their keys arrive as install inputs.
          </Text>
        </div>
      ) : null}
    </div>
  )
}

// Manual setup, push-based. The config lives in the repo connected in Set up and the
// default app branch tracks it, so a push is the sync (docs/guides/app-branches: any
// push to the tracked branch starts a run). The six files live here, for the person
// who chose to fill them in.
const ManualSetup = ({ appName, repo, cloud }: { appName: string; repo: string; cloud: TCloud }) => {
  const steps: { title: string; body: ReactNode; detail?: ReactNode }[] = [
    {
      title: 'Put the config at the root of the repo',
      body: (
        <>
          Top level of <Badge size="sm" variant="code">{repo}</Badge>, the same layout as{' '}
          <Link href={KITCHEN_SINK_REPO} isExternal textVariant="subtext" className="!inline-flex align-baseline">
            nuonco/kitchen-sink
          </Link>
        </>
      ),
    },
    {
      title: 'Fill in the app config templates with your values',
      body: (
        <>
          Point each component at a repo and branch, pick a sandbox, scope roles.{' '}
          <Link href={DOCS_CONFIG_FILES} isExternal textVariant="subtext" className="!inline-flex align-baseline">
            Configuration files
          </Link>
        </>
      ),
      detail: <FileStubRows appName={appName} cloud={cloud} />,
    },
    {
      title: 'Push to main',
      body: <>Every push syncs the default app branch.</>,
      detail: (
        <CodeBlock language="bash" showCopy>
          {GIT_PUSH(appName)}
        </CodeBlock>
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
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
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
      <Text variant="subtext" theme="neutral" flex className="flex-wrap border-t pt-3">
        Editing TOML by hand? The Nuon language server adds autocomplete and validation:
        <Link href={VSCODE_EXTENSION} isExternal textVariant="subtext">
          VS Code extension
        </Link>
        <span aria-hidden>·</span>
        <Link href={LSP_NEOVIM_SETUP} isExternal textVariant="subtext">
          Neovim setup
        </Link>
        <span aria-hidden>·</span>
        <Link href={DOCS_LSP} isExternal textVariant="subtext">
          Language server docs
        </Link>
      </Text>
    </div>
  )
}

// The agent path is the card. The prompt is the one thing to act on.
const AgentSetup = () => (
  <div className="flex flex-col gap-4">
    <div className="flex flex-col gap-1.5">
      <Text as="h3" variant="h3" weight="strong" flex>
        <Icon variant="RobotIcon" size={20} />
        Have your agent write the config
      </Text>
      <Text variant="body" theme="neutral">
        Paste this prompt in the same directory as your app. Then, you&apos;re one step from a test install as if
        it were a customer&apos;s cloud.
      </Text>
    </div>
    <div className="flex flex-col gap-4 rounded-md border bg-background p-4 sm:flex-row sm:items-center">
      <div className="line-clamp-1 min-w-0 flex-1">
        <Text as="span" variant="body" family="mono" weight="strong" theme="brand">
          /goal
        </Text>
        <Text as="span" variant="body" family="mono" theme="neutral">
          {AGENT_PASTE.slice(5)}
        </Text>
      </div>
      <div className="flex shrink-0 flex-wrap items-center gap-2">
        <CopyTextButton text={AGENT_PASTE} label="Copy prompt" size="lg" variant="primary" />
        <Button variant="secondary" size="lg" href={PROMPT_TXT_URL} target="_blank" rel="noreferrer">
          See full prompt <Icon variant="ArrowSquareOutIcon" size={14} />
        </Button>
      </div>
    </div>
  </div>
)

const TemplateStep = ({ sharedData, setSharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const { choose, pushTick } = useForkChoice()
  const appName = readAppName(sharedData)
  const cloud = readCloud(sharedData)
  // The connected account from Set up, and a repo named after the app.
  const repo = `jane-doe/${appName}`
  const detected = pushTick > 0

  // Back to the fork, collapsed, with the example path selected.
  const exitToExample = () => {
    setSharedData('expandOwn', false)
    setSharedData('path', 'example')
    setSharedData('cloud', EXAMPLE_CLOUDS[0])
    setSharedData('region', CLOUD_REGIONS[EXAMPLE_CLOUDS[0]].options[0])
    choose({ path: 'example', cloud: EXAMPLE_CLOUDS[0] })
    onGoBack?.()
  }

  return (
    <div className="flex flex-col gap-6">
      <Card className="!gap-10 !p-5 !border-0 !shadow-none bg-primary-50 dark:bg-primary-950/40 ring-1 ring-primary-200 dark:ring-primary-800">
        <AgentSetup />
        <PushListener detected={detected} cloud={cloud} />
      </Card>
      <Footnotes appName={appName} repo={repo} cloud={cloud} onExampleExit={exitToExample} />
      <NextButton label="Set up your first install" onClick={onAdvance} onBack={onGoBack} />
    </div>
  )
}

const ForkStep = ({ sharedData, setSharedData, onAdvance }: IWizardStepComponentProps) => {
  const { choose, backToIntro } = useForkChoice()
  const [expanded, setExpanded] = useState(Boolean(sharedData.expandOwn))
  const setupRef = useRef<HTMLDivElement>(null)
  const named = ((sharedData.appName as string | undefined) ?? '').trim().length > 0
  const githubDone = Boolean(sharedData.githubDone)
  const testCloud = sharedData.testCloud as TCloud | undefined
  // Errors show only after a failed attempt to continue, on whichever field is missing.
  const [showErrors, setShowErrors] = useState(false)

  const tryContinue = () => {
    if (!named || !githubDone || !testCloud) {
      setShowErrors(true)
      setupRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
      return
    }
    go({ path: 'own', cloud: testCloud })
  }

  const go = (choice: IForkChoice) => {
    choose(choice)
    setSharedData('path', choice.path)
    setSharedData('cloud', choice.cloud ?? 'aws')
    onAdvance()
  }

  // Expanding commits to the own-app path so the stepper stops showing the example path's "Deploy" dot.
  const expand = () => {
    choose({ path: 'own', cloud: readCloud(sharedData) })
    setSharedData('path', 'own')
    // Persisted so Back from the template step remounts this step still expanded.
    setSharedData('expandOwn', true)
    setExpanded(true)
    requestAnimationFrame(() => setupRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
  }

  const exitToExample = () => {
    setSharedData('expandOwn', false)
    setSharedData('path', 'example')
    setSharedData('cloud', EXAMPLE_CLOUDS[0])
    setSharedData('region', CLOUD_REGIONS[EXAMPLE_CLOUDS[0]].options[0])
    choose({ path: 'example', cloud: EXAMPLE_CLOUDS[0] })
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
            cloud={testCloud}
            onCloud={(value) => {
              setSharedData('testCloud', value)
              setSharedData('cloud', value)
              setSharedData('region', CLOUD_REGIONS[value].options[0])
              choose({ path: 'own', cloud: value })
            }}
          />
          <ExampleEscapeHatch onExit={exitToExample} />
        </div>
      ) : (
        <>
        <Card>
          {heading}
          <ol className="grid gap-3 sm:grid-cols-3">
              {OWN_APP_STEPS.map((step, index) => (
                <li
                  key={step.title}
                  className="flex items-center gap-2.5 rounded-md border px-3.5 py-3"
                >
                  <Icon variant={step.icon} size={18} theme="brand" />
                  <Text variant="body" weight="strong" className="min-w-0 flex-1">
                    {step.title}
                  </Text>
                  <Badge size="sm" theme="brand">
                    {index + 1}
                  </Badge>
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
            <Icon variant="TireIcon" size={20} theme="neutral" />
            <Text variant="h3" role="heading" level={3}>
              Or kick the tires with our example app first
            </Text>
          </div>
          <Text variant="body" theme="neutral">
            Pre-wired with Terraform, Helm, images, manifests. Deploy it to your cloud account just
            like your customers would deploy your app.
          </Text>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          {EXAMPLE_CLOUDS.map((cloud) => (
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
        </div>
        <ExampleAppDrawer />
      </Card>
        </>
      )}


      {expanded ? (
        <NextButton label="Connect your app" onClick={tryContinue} onBack={backToIntro} />
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

const TEST_CLOUDS: TCloud[] = ['aws', 'gcp', 'azure']

// No default: the test cloud is the user's own account, so nothing is preselected.
const TestCloudPicker = ({
  value,
  onChange,
  error,
}: {
  value?: TCloud
  onChange: (cloud: TCloud) => void
  error: boolean
}) => (
  <fieldset aria-describedby={error ? 'test-cloud-error' : 'test-cloud-hint'}>
    <legend className="mb-2">
      <Text variant="body" weight="strong">
        Test cloud where your app will be installed
      </Text>
    </legend>
    <div className="flex flex-col gap-2">
      <div className="grid max-w-md grid-cols-3 gap-3">
        {TEST_CLOUDS.map((cloud) => {
          const checked = value === cloud
          return (
            <label
              key={cloud}
              title={CLOUD_LABEL[cloud]}
              className={cn(
                'flex h-14 cursor-pointer items-center gap-3 rounded-md px-4 ring-1 transition-shadow',
                'has-[:focus-visible]:outline has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-primary-500',
                checked
                  ? 'ring-2 ring-primary-500 bg-primary-50 dark:bg-primary-950/40'
                  : error
                    ? 'ring-red-500 dark:ring-red-400 hover:bg-neutral-50 dark:hover:bg-neutral-900'
                    : 'ring-neutral-200 dark:ring-neutral-700 hover:bg-neutral-50 dark:hover:bg-neutral-900'
              )}
            >
              <input
                type="radio"
                name="test-cloud"
                value={cloud}
                checked={checked}
                onChange={() => onChange(cloud)}
                required
                aria-invalid={error || undefined}
                className="accent-primary-600 focus-visible:outline-none"
              />
              <span className="flex flex-1 justify-center">
                <Icon variant={CLOUD_ICON[cloud]} size={cloud === 'aws' ? 26 : 22} />
              </span>
              <span className="sr-only">{CLOUD_LABEL[cloud]}</span>
            </label>
          )
        })}
      </div>
      {error ? (
        <Text id="test-cloud-error" variant="subtext" theme="error" flex>
          <Icon variant="WarningCircleIcon" size={14} weight="fill" />
          Select a test cloud to continue.
        </Text>
      ) : (
        <Text id="test-cloud-hint" variant="subtext" theme="neutral">
          Nuon stubs the runner, sandbox and permissions for this cloud.
        </Text>
      )}
    </div>
  </fieldset>
)

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
    { name: 'Azure CLI (Bicep)', how: 'Create a resource group and Key Vault, then deploy the template with az. Nuon fills in the commands.' },
    { name: 'Terraform', how: 'Generated tfvars for the install-stacks/azure module, applied with terraform.' },
    { name: 'Deploy to Azure', how: 'A pre-filled portal link. Only when stack.toml sets deployment_scope = "subscription".' },
  ],
}

// What this install will contain, as a card worth reading: source, sandbox, and
// components, plus the same three tiers the intro drew. Framing-agnostic — the
// example and own paths differ only in the facts.
// Both repo facts are chips that open the repo.
const RepoChip = ({ repo }: { repo: string }) => (
  <Link href={`https://github.com/${repo}`} isExternal textVariant="subtext">
    <Badge size="sm" variant="code">
      {repo}
    </Badge>
  </Link>
)

const InstallSummaryCard = ({
  path,
  appName,
  cloud,
}: {
  path: TPath
  appName: string
  cloud: TCloud
}) => {
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
        {
          label: 'Nuon sandbox',
          value: (
            <>
              <RepoChip repo={CLOUD_SANDBOX[cloud]} />
              <Text variant="subtext" theme="neutral">
                from sandbox.toml
              </Text>
            </>
          ),
        },
        { label: 'Components', value: 'api: Helm chart, from components/api.toml' },
      ]
    : [
        { label: 'Source', value: <RepoChip repo="nuonco/kitchen-sink" /> },
        { label: 'Nuon sandbox', value: <RepoChip repo={CLOUD_SANDBOX[cloud]} /> },
        { label: 'Components', value: 'Terraform modules, Helm charts, container images' },
      ]

  return (
    <Card className="!gap-5">
      <div className="flex items-center gap-3">
        <Icon variant={own ? 'GitBranchIcon' : 'TireIcon'} size={24} theme="brand" />
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
  const region = (sharedData.region as string | undefined) ?? regions.options[0]
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
        <Select
          id="fork-region"
          options={regionOptions(cloud)}
          labelProps={{ labelText: regions.label }}
          value={region}
          onChange={(value) => setSharedData('region', value)}
        />
        <Toggle
          checked={autoApprove}
          onChange={setAutoApprove}
          label="Auto-approve"
          description="Applies each plan as soon as it is ready. On by default for a faster first run."
        />
      </Card>

      <InstallSummaryCard path={path} appName={appName} cloud={cloud} />

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
  const region = (sharedData.region as string | undefined) ?? regions.options[0]
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
    ? `Generating the ${connect.artifactNoun} for ${region}. About 30 seconds.`
    : ready
      ? `${connect.artifactNoun} ready for ${region}. From launch to a healthy runner is about 11 minutes. This page updates on its own.`
      : phase === 'waiting'
        ? `${connect.waitingHint} This page updates on its own.`
        : phase === 'done'
          ? `${connect.stackLabel} created. Test ${connect.accountNoun} connected.`
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
              Test {connect.accountNoun} · {region}
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
            {cloud === 'gcp'
              ? 'On Google Cloud, Nuon renders the install stack in Terraform.'
              : `Nuon renders the install stack in Terraform and in ${CLOUD_LABEL[cloud]}'s native format.`}{' '}
            Your customer creates it with their own credentials; that is how access is granted. You are
            about to do it the way they would.
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
            generating ? { tipContent: `Cannot launch until Nuon finishes generating the ${connect.artifactNoun}` } : undefined
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

type TStageId = 'runner' | 'sandbox' | 'components'
type TStageState = 'done' | 'active' | 'next'

interface IBuildStage {
  id: TStageId
  label: string
  duration: string
  activeStatus: string
  blurb: string
}

// Onboarding is a proof of concept, so the account is always framed as a test one.
const accountLabel = (_path: TPath, cloud: TCloud) => `your test ${CLOUD_CONNECT[cloud].accountNoun}`

const SANDBOX_CLUSTER: Record<TCloud, string> = {
  aws: 'An EKS cluster and node group',
  gcp: 'A GKE Autopilot cluster',
  azure: 'An AKS cluster',
}
const SANDBOX_PARTS = ['cluster', 'registry', 'ingress', 'namespaces']

// What is left AFTER the stack, which the user created on the previous step.
// Durations from docs/get-started: a healthy runner ≈ 1 min after the stack
// reports home; eks-simple end to end ≈ 35 min.
const buildStages = (path: TPath, cloud: TCloud, appName: string): IBuildStage[] => [
  {
    id: 'runner',
    label: 'Nuon runner',
    duration: 'about 1 min',
    activeStatus: 'Starting',
    blurb: `Runs in ${accountLabel(path, cloud)} and builds everything else, using the roles your stack granted.`,
  },
  {
    id: 'sandbox',
    label: 'Nuon sandbox',
    duration: 'about 15–20 min',
    activeStatus: 'Creating',
    blurb: `${SANDBOX_CLUSTER[cloud]}, a container registry, ingress and namespaces. Takes about 15–20 minutes.`,
  },
  {
    id: 'components',
    label: 'Components',
    duration: 'a few min',
    activeStatus: 'Deploying',
    blurb: `${appName}'s Terraform, Helm charts and images, deployed into the sandbox. A few minutes.`,
  },
]

const stageState = (index: number, activeIndex: number): TStageState =>
  index < activeIndex ? 'done' : index === activeIndex ? 'active' : 'next'

// Placeholder in Nuon's install-ID shape; the product passes the real one.
const EXAMPLE_INSTALL_ID = 'inlk3x9q2m7v4w8p1z6r5t0y2c'

const ProvisionAccountView = ({
  stages,
  activeIndex,
  cloud,
  region,
}: {
  stages: IBuildStage[]
  activeIndex: number
  cloud: TCloud
  region: string
}) => {
  const [picked, setPicked] = useState<TStageId | null>(null)
  const [hovered, setHovered] = useState<TStageId | null>(null)
  const running = stages[Math.min(activeIndex, stages.length - 1)].id
  const focus = hovered ?? picked ?? running
  const stateOf = (id: TStageId) => stageState(stages.findIndex((stage) => stage.id === id), activeIndex)
  const region_ = (id: TStageId) =>
    focus === id
      ? 'ring-2 ring-primary-500 bg-primary-50 dark:bg-primary-950/40'
      : 'ring-1 ring-neutral-200 dark:ring-neutral-700 bg-background'
  const runner = stateOf('runner')
  const sandbox = stateOf('sandbox')
  const components = stateOf('components')

  return (
    <div className="flex flex-col gap-5 md:flex-row">
      <ol className="flex shrink-0 flex-col gap-2 md:w-60" aria-label="Install stages">
        {stages.map((stage, index) => {
          const state = stageState(index, activeIndex)
          const on = focus === stage.id
          return (
            <li key={stage.id}>
              <button
                type="button"
                aria-pressed={on}
                onClick={() => setPicked(stage.id)}
                onMouseEnter={() => setHovered(stage.id)}
                onMouseLeave={() => setHovered(null)}
                onFocus={() => setHovered(stage.id)}
                onBlur={() => setHovered(null)}
                className={cn(
                  'flex w-full flex-col gap-1.5 rounded-lg px-3 py-2.5 text-left transition-colors',
                  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-500',
                  on
                    ? 'ring-1 ring-primary-500 bg-primary-50 dark:bg-primary-950/40'
                    : 'ring-1 ring-neutral-200 dark:ring-neutral-700 hover:bg-neutral-50 dark:hover:bg-neutral-900'
                )}
              >
                <span className="flex items-center gap-2.5">
                  <span
                    className={cn(
                      'flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] font-semibold',
                      state === 'done'
                        ? 'bg-green-600 text-white'
                        : state === 'active'
                          ? 'bg-primary-50 text-primary-800 ring-2 ring-primary-500 dark:bg-primary-950'
                          : 'bg-background text-neutral-500 ring-2 ring-neutral-200 dark:ring-neutral-700'
                    )}
                  >
                    {state === 'done' ? <Icon variant="CheckIcon" size={12} weight="bold" /> : index + 1}
                  </span>
                  <Text variant="body" weight="strong" className="min-w-0 flex-1">
                    {stage.label}
                  </Text>
                  <Text
                    variant="label"
                    weight="strong"
                    theme={state === 'done' ? 'success' : state === 'active' ? 'brand' : 'neutral'}
                  >
                    {state === 'done' ? 'Done' : state === 'active' ? stage.activeStatus : 'Up next'}
                  </Text>
                </span>
                {on ? (
                  <Text variant="subtext" theme="neutral" className="pl-8">
                    {stage.blurb}
                  </Text>
                ) : null}
              </button>
            </li>
          )
        })}
      </ol>

      <div className="flex min-w-0 flex-1 flex-col gap-3 rounded-xl p-4 ring-2 ring-primary-500 bg-primary-50/40 dark:bg-primary-950/20">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Icon variant={CLOUD_ICON[cloud]} size={18} />
            <Text variant="base" weight="strong">
              Your test {CLOUD_CONNECT[cloud].accountNoun}
            </Text>
          </div>
          <Badge size="sm" variant="code">
            {region}
          </Badge>
        </div>

        <div className={cn('flex items-center gap-3 rounded-lg px-3.5 py-3 transition-colors', region_('runner'))}>
          {runner === 'done' ? (
            <Icon variant="CheckCircleIcon" size={22} weight="fill" theme="success" />
          ) : (
            <Icon variant="Loading" size={20} />
          )}
          <div className="flex min-w-0 flex-col">
            <Text variant="body" weight="strong">
              Nuon runner
            </Text>
            <Text variant="subtext" theme="neutral">
              {runner === 'done'
                ? 'Running. Everything below is built by it, from inside your account.'
                : 'Starting on the machine your stack created.'}
            </Text>
          </div>
        </div>

        <div className={cn('flex flex-col gap-3 rounded-lg p-3.5 transition-colors', region_('sandbox'))}>
          <div className="flex items-center justify-between gap-3">
            <Text variant="body" weight="strong">
              Nuon sandbox
            </Text>
            <Text variant="subtext" weight="strong" theme={sandbox === 'done' ? 'success' : sandbox === 'active' ? 'brand' : 'neutral'}>
              {sandbox === 'done' ? 'Ready' : sandbox === 'active' ? 'Creating · 15–20 min' : 'Up next · 15–20 min'}
            </Text>
          </div>
          <div className="flex flex-wrap gap-2">
            {SANDBOX_PARTS.map((part, index) => {
              const built = sandbox === 'done'
              const building = sandbox === 'active' && index === 0
              return (
                <span
                  key={part}
                  className={cn(
                    'rounded-md px-2 py-1 font-mono text-xs',
                    built
                      ? 'ring-1 ring-neutral-300 dark:ring-neutral-600 bg-background'
                      : building
                        ? 'ring-1 ring-primary-500 bg-primary-50 dark:bg-primary-950/40'
                        : 'outline-1 outline-dashed outline-neutral-300 dark:outline-neutral-600 text-neutral-500'
                  )}
                >
                  {part}
                </span>
              )
            })}
          </div>
          <div
            className={cn(
              'flex items-center justify-between gap-3 rounded-lg px-3.5 py-3 transition-colors',
              focus === 'components'
                ? 'ring-2 ring-primary-500 bg-primary-50 dark:bg-primary-950/40'
                : components === 'next'
                  ? 'outline-1 outline-dashed outline-neutral-300 dark:outline-neutral-600 bg-background'
                  : 'ring-1 ring-neutral-200 dark:ring-neutral-700 bg-background'
            )}
          >
            <div className="flex min-w-0 flex-col">
              <Text variant="body" weight="strong">
                Your components
              </Text>
              <Text variant="subtext" theme="neutral">
                Terraform, Helm charts and images land here once the sandbox is up.
              </Text>
            </div>
            <Text
              variant="subtext"
              weight="strong"
              theme={components === 'done' ? 'success' : components === 'active' ? 'brand' : 'neutral'}
              className="whitespace-nowrap"
            >
              {components === 'done' ? 'Deployed' : components === 'active' ? 'Deploying' : 'Up next'}
            </Text>
          </div>
        </div>
      </div>
    </div>
  )
}

const STAGE_DWELL_MS = [3000, 6000, 3000]

const ProvisionStep = ({ sharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const path = readPath(sharedData)
  const cloud = readCloud(sharedData)
  const appName = path === 'own' ? readAppName(sharedData) : 'Kitchen Sink'
  const stages = buildStages(path, cloud, appName)
  const region = (sharedData.region as string | undefined) ?? CLOUD_REGIONS[cloud].options[0]
  const [activeIndex, setActiveIndex] = useState(0)
  const watchCommand = `nuon installs workflows watch -i ${EXAMPLE_INSTALL_ID}`

  useEffect(() => {
    if (activeIndex >= stages.length) return
    const timer = setTimeout(() => setActiveIndex((prev) => prev + 1), STAGE_DWELL_MS[activeIndex] ?? 3000)
    return () => clearTimeout(timer)
  }, [activeIndex, stages.length])

  return (
    <div className="flex flex-col gap-6">
      <Card className="!gap-5">
        <ProvisionAccountView stages={stages} activeIndex={activeIndex} cloud={cloud} region={region} />
        <div className="flex flex-col gap-3 border-t pt-5">
          <div className="flex flex-col gap-1">
            <Text variant="body" weight="strong">
              See it in action in the CLI
            </Text>
            <Text variant="subtext" theme="neutral" flex>
              Needs the Nuon CLI:
              <Badge size="sm" variant="code">
                brew install nuonco/tap/nuon
              </Badge>
              then
              <Badge size="sm" variant="code">
                nuon login
              </Badge>
            </Text>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Badge size="sm" variant="code">
              {watchCommand}
            </Badge>
            <CopyTextButton text={watchCommand} label="Copy" size="sm" />
          </div>
        </div>
      </Card>

      <NextButton label="Go to deploy workflow" onClick={onAdvance} onBack={onGoBack} />
    </div>
  )
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
  title: 'Connect your app',
  navLabel: 'Connect',
  description: "This config is how Nuon installs and upgrades your app in every customer's cloud.",
  component: TemplateStep,
}

const DEPLOY_STEP: IWizardStepDef = {
  id: 'deploy',
  title: 'Your app is ready for BYOC',
  navLabel: 'Deploy',
  description:
    'Now you can test the flow your customer will see. Pick a cloud account you want to test with.',
  component: DeployStep,
}

const STACK_STEP_INTRO: Record<TCloud, ReactNode> = {
  aws: (
    <>
      Nuon is generating a{' '}
      <Link href={AWS_QUICK_CREATE_DOCS} isExternal textVariant="body" className="!inline-flex align-baseline">
        CloudFormation quick-create link
      </Link>
      . This is a common install method for BYOC customers on AWS.
    </>
  ),
  gcp: (
    <>
      Nuon is generating the Terraform for your stack. On Google Cloud, BYOC customers apply it
      themselves or through{' '}
      <Link href={GCP_INFRA_MANAGER_DOCS} isExternal textVariant="body" className="!inline-flex align-baseline">
        Infrastructure Manager
      </Link>
      .
    </>
  ),
  azure: (
    <>
      Nuon is generating the Bicep template and the commands that deploy it. At the default{' '}
      <Link href={DOCS_STACKS} isExternal textVariant="body" className="!inline-flex align-baseline">
        resource group scope
      </Link>
      , BYOC customers create a resource group and Key Vault first, then run the commands.
    </>
  ),
}

const INSTALL_STACK_STEP: IWizardStepDef = {
  id: 'install-stack',
  title: 'Create the install stack',
  navLabel: 'Stack',
  description: STACK_STEP_INTRO.aws,
  component: StackStep,
}

const PROVISION_STEP: Record<TPath, IWizardStepDef> = {
  example: {
    id: 'example-provision',
    title: 'Your first BYOC install is deploying',
    navLabel: 'Provision',
    component: ProvisionStep,
  },
  own: {
    id: 'own-provision',
    title: 'Your first BYOC install is deploying',
    navLabel: 'Provision',
    component: ProvisionStep,
  },
}

const buildForkFlow = (path: TPath, cloud: TCloud): IWizardStepDef[] => {
  const stackStep = { ...INSTALL_STACK_STEP, description: STACK_STEP_INTRO[cloud] }
  if (path === 'own') return [FORK_STEP, TEMPLATE_STEP, DEPLOY_STEP, stackStep, PROVISION_STEP.own]
  return [
    FORK_STEP,
    { ...DEPLOY_STEP, id: `deploy-${cloud}` },
    { ...stackStep, id: `install-stack-${cloud}` },
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

  const fork: IForkActions = { choose: setChoice, backToIntro: () => reset(true), pushTick: pushes }

  if (finished) {
    return <FlowComplete onRestart={() => reset(initialStepIndex > 0 || skipIntro)} />
  }

  return (
    <CopyEditor
      tools={
        <div className="flex items-center gap-2">
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


export const ForkOwnApp = () => <BranchingPlayground initialPath="own" skipIntro expandOwnApp />
ForkOwnApp.meta = { fullBleed: true }
