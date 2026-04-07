export default {
  title: 'Onboarding/OnboardingWizard',
}

import { WizardContext } from '@/providers/onboarding-wizard-provider'
import { ConfigContext } from '@/providers/config-provider'
import type { IWizardContext } from '@/providers/onboarding-wizard-provider'
import type { TRuntimeConfig } from '@/providers/config-provider'
import { Text } from '@/components/common/Text'
import { OnboardingWizardLayout } from './OnboardingWizard'

const MockStep = () => (
  <div className="flex flex-col gap-4 p-4 border rounded">
    <Text variant="body">This is a mock wizard step.</Text>
  </div>
)

const mockSteps = [
  { id: 'welcome', title: 'Welcome', description: 'Get started with Nuon.', component: MockStep },
  { id: 'create-app', title: 'Create your app', description: 'Set up your first app.', component: MockStep },
  { id: 'sync', title: 'Sync app', description: 'Sync your app config.', component: MockStep },
]

const mockWizard: IWizardContext = {
  steps: mockSteps,
  currentStepIndex: 0,
  completedSteps: new Set<string>(),
  sharedData: {},
  canClose: true,
  markComplete: () => {},
  setSharedData: () => {},
  goToStep: () => {},
  goNext: () => {},
  goPrev: () => {},
  onComplete: () => {},
}

const mockConfig: TRuntimeConfig = {
  apiUrl: 'http://localhost:8081',
  appUrl: 'http://localhost:4000',
  githubAppName: 'nuon-dev',
  isByoc: false,
  onboardingV2: false,
}

const Providers = ({ children, wizard = mockWizard, config = mockConfig }: {
  children: React.ReactNode
  wizard?: IWizardContext
  config?: TRuntimeConfig
}) => (
  <ConfigContext.Provider value={config}>
    <WizardContext.Provider value={wizard}>
      {children}
    </WizardContext.Provider>
  </ConfigContext.Provider>
)

export const Default = () => (
  <Providers>
    <OnboardingWizardLayout onboardingV2={false} skipHref="/org-123/apps" />
  </Providers>
)

export const V2 = () => (
  <Providers config={{ ...mockConfig, onboardingV2: true }}>
    <OnboardingWizardLayout onboardingV2={true} skipHref={null} />
  </Providers>
)

export const MidProgress = () => (
  <Providers wizard={{ ...mockWizard, currentStepIndex: 1, completedSteps: new Set(['welcome']) }}>
    <OnboardingWizardLayout onboardingV2={false} skipHref="/org-123/apps" />
  </Providers>
)
