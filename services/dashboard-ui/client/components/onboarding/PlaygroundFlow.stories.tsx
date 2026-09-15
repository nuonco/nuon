export default {
  title: 'Onboarding/Playground',
}

import { useEffect, useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
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
  onClick,
}: {
  label?: string
  disabled?: boolean
  onClick: () => void
}) => (
  <div className="flex justify-end">
    <Button variant="primary" disabled={disabled} onClick={onClick}>
      {label ?? 'Continue'} <Icon variant="CaretRightIcon" weight="bold" />
    </Button>
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
