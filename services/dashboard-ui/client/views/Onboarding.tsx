import { OnboardingWizard } from '@/components/onboarding/OnboardingWizard'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

const PlaceholderStep = ({ onAdvance }: IWizardStepComponentProps) => {
  return (
    <div className="py-12 items-center flex flex-col gap-4">
      <Text variant="body">Step content goes here.</Text>
      <Button onClick={onAdvance} variant="primary">
        Complete step
      </Button>
    </div>
  )
}

const STEPS = [
  {
    id: 'step-1',
    title: 'Welcome to Nuon',
    description: "Let's get you set up.",
    component: PlaceholderStep,
  },
  {
    id: 'step-2',
    title: 'Create your org',
    description: 'Set up your organization.',
    component: PlaceholderStep,
  },
  {
    id: 'step-3',
    title: 'Download the Nuon CLI',
    description: 'Lets download the Nuon CLI.',
    component: PlaceholderStep,
  },
  {
    id: 'step-4',
    title: 'Create your first app',
    description: 'Choose an example app to get started.',
    component: PlaceholderStep,
  },
  {
    id: 'step-5',
    title: 'Sync your app config',
    description: 'Sync your example app config to get ready for deployment.',
    component: PlaceholderStep,
  },
  {
    id: 'step-6',
    title: 'Create an install',
    description: 'Create an install of your app config.',
    component: PlaceholderStep,
  },
]

export function Onboarding() {
  return (
    <OnboardingWizard
      steps={STEPS}
      onComplete={() => {
        window.location.href = '/'
      }}
    />
  )
}
