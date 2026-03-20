import { OnboardingWizard } from '@/components/onboarding/OnboardingWizard'
import { WelcomeStep } from '@/components/onboarding/steps/WelcomeStep'
import { CreateOrgStep } from '@/components/onboarding/steps/CreateOrgStep'
import { DownloadCliStep } from '@/components/onboarding/steps/DownloadCliStep'
import { CreateAppStep } from '@/components/onboarding/steps/CreateAppStep'
import { SyncAppStep } from '@/components/onboarding/steps/SyncAppStep'
import { CreateInstallStep } from '@/components/onboarding/steps/CreateInstallStep'
import { WelcomeNameOrgStep } from '@/components/onboarding/steps/v2/WelcomeNameOrgStep'
import { AppProfileStep } from '@/components/onboarding/steps/v2/AppProfileStep'
import { CloudSetupStep } from '@/components/onboarding/steps/v2/CloudSetupStep'
import { ProvisioningStep } from '@/components/onboarding/steps/v2/ProvisioningStep'
import { NextStepsStep } from '@/components/onboarding/steps/v2/NextStepsStep'
import { OnboardingJourneyProvider } from '@/providers/onboarding-journey-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import { useConfig } from '@/hooks/use-config'

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

const STEPS_V2 = [
  {
    id: 'v2-step-1',
    title: 'Welcome to Nuon',
    navLabel: 'Welcome',
    component: WelcomeNameOrgStep,
  },
  {
    id: 'v2-step-2',
    title: 'Your app profile',
    navLabel: 'App profile',
    component: AppProfileStep,
  },
  {
    id: 'v2-step-3',
    title: 'Connect your cloud account',
    description: "Connect a runner to deploy into a real cloud environment, or let Nuon spin up a sandbox so you can explore the platform first.",
    navLabel: 'Cloud setup',
    component: CloudSetupStep,
  },
  {
    id: 'v2-step-4',
    title: 'Provisioning',
    navLabel: 'Provisioning',
    component: ProvisioningStep,
  },
  {
    id: 'v2-step-5',
    title: 'Next steps',
    navLabel: 'Next steps',
    component: NextStepsStep,
  },
]

export function Onboarding() {
  const { onboardingV2 } = useConfig()
  const steps = onboardingV2 ? STEPS_V2 : STEPS

  return (
    <ToastProvider>
      <SurfacesProvider>
        <OnboardingJourneyProvider>
          <OnboardingWizard
            steps={steps}
            onComplete={() => {
              window.location.href = '/'
            }}
          />
        </OnboardingJourneyProvider>
      </SurfacesProvider>
    </ToastProvider>
  )
}
