import { OnboardingWizard } from '@/components/onboarding/OnboardingWizard'
import { WelcomeStep } from '@/components/onboarding/steps/WelcomeStep'
import { CreateOrgStep } from '@/components/onboarding/steps/CreateOrgStep'
import { DownloadCliStep } from '@/components/onboarding/steps/DownloadCliStep'
import { CreateAppStep } from '@/components/onboarding/steps/CreateAppStep'
import { SyncAppStep } from '@/components/onboarding/steps/SyncAppStep'
import { CreateInstallStep } from '@/components/onboarding/steps/CreateInstallStep'
import { OnboardingJourneyProvider } from '@/providers/onboarding-journey-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'

const STEPS = [
  {
    id: 'step-1',
    title: 'Welcome to Nuon',
    navLabel: 'Get Started',
    description: "Let's setup your account.",
    component: WelcomeStep,
  },
  {
    id: 'step-2',
    title: 'Create your org',
    navLabel: 'Create Org',
    description: 'Set up your organization.',
    component: CreateOrgStep,
  },
  {
    id: 'step-3',
    title: 'Download the Nuon CLI',
    navLabel: 'Nuon CLI',
    description:
      'Download the Nuon CLI to create and manage your apps from the terminal.',
    component: DownloadCliStep,
  },
  {
    id: 'step-4',
    title: 'Create your first app',
    navLabel: 'Create App',
    description:
      'Choose an example app to get started. You can customize it later.',
    component: CreateAppStep,
  },
  {
    id: 'step-5',
    title: 'Sync your app config',
    navLabel: 'Sync App',
    description:
      'Syncing pushes your app config to Nuon and triggers a build. Run this from inside your cloned app directory.',
    component: SyncAppStep,
  },
  {
    id: 'step-6',
    title: 'Create an install',
    navLabel: 'Deploy Install',
    description: 'Create an install to deploy your app to a cloud account.',
    component: CreateInstallStep,
  },
]

export function Onboarding() {
  return (
    <ToastProvider>
      <SurfacesProvider>
        <OnboardingJourneyProvider>
          <OnboardingWizard
            steps={STEPS}
            onComplete={() => {
              window.location.href = '/'
            }}
          />
        </OnboardingJourneyProvider>
      </SurfacesProvider>
    </ToastProvider>
  )
}
