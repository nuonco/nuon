import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router'
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
import { createOnboarding, completeOrganizationStep } from '@/lib'
import type { TOnboarding } from '@/types'

const ONBOARDING_STEP_TO_INDEX: Record<string, number> = {
  organization: 0,
  your_stack: 1,
  install: 2,
  deploy: 3,
  get_started: 4,
}

const STEPS = [
  {
    id: 'step-1',
    title: 'Welcome to Nuon',
    navLabel: 'Get Started',
    component: WelcomeStep,
  },
  {
    id: 'step-2',
    title: 'Create your org',
    navLabel: 'Create Org',
    description: (
      <div className="flex flex-col gap-2">
        <p>
          An org is an isolated place for metadata about your apps, installs, workflows and logs in Nuon Cloud.
        </p>
        <p className="text-sm opacity-80">
          <a
            href="https://nuon.co/contact-sales"
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary-600 dark:text-primary-400 underline underline-offset-2"
          >
            Contact sales
          </a>{' '}
          if you want to run Nuon&apos;s control plane in your AWS, Azure, or GCP.
        </p>
      </div>
    ),
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
    title: 'Sync your app',
    navLabel: 'Sync App',
    description: (
      <div className="flex flex-col gap-2">
        <p>
          Syncing pushes your app to Nuon and triggers a build. Run this from inside your cloned app directory.
        </p>
        <p className="text-sm opacity-80">
          A build creates OCI artifacts for the components in your app (e.g., Helm, Terraform, container image, Kubernetes manifest, etc.) and stores them in an isolated container registry in your org.
        </p>
      </div>
    ),
    component: SyncAppStep,
  },
  {
    id: 'step-6',
    title: 'Create an install',
    navLabel: 'Deploy Install',
    description: 'Create an install to deploy your app to a cloud account. (Scroll down to find the Create install button.)',
    component: CreateInstallStep,
  },
]

const STEPS_V2 = [
  {
    id: 'v2-step-1',
    title: 'Create your organization',
    navLabel: 'Organization',
    hideTitle: true,
    component: WelcomeNameOrgStep,
  },
  {
    id: 'v2-step-2',
    title: 'Tell us about your app',
    navLabel: 'Your stack',
    description: 'Pick your cloud platform and app attributes, or start from a working example.',
    component: AppProfileStep,
  },
  {
    id: 'v2-step-3',
    title: 'Choose how to deploy',
    navLabel: 'Install',
    description: 'Connect your own cloud account or use a managed sandbox to explore the platform.',
    component: CloudSetupStep,
  },
  {
    id: 'v2-step-4',
    title: 'Your install is being created!',
    navLabel: 'Deploy',
    description: 'Hang tight. While the resources are getting provisioned.',
    component: ProvisioningStep,
  },
  {
    id: 'v2-step-5',
    title: "You're all set",
    navLabel: 'Get started',
    hideTitle: true,
    component: NextStepsStep,
  },
]

export function Onboarding() {
  const { onboardingV2 } = useConfig()
  const [searchParams] = useSearchParams()
  const requestedOrgId = searchParams.get('org_id')
  const steps = onboardingV2 ? STEPS_V2 : STEPS
  const [initialSharedData, setInitialSharedData] = useState<
    Record<string, unknown> | undefined
  >(onboardingV2 ? undefined : {})

  useEffect(() => {
    if (!onboardingV2) return
    let cancelled = false
    const load = async () => {
      try {
        let ob = await createOnboarding()
        // Auto-attach the org passed in via `?org_id=` whenever it differs
        // from whatever is currently associated with the onboarding session.
        // The backend rejects switching after resources have been created,
        // in which case we just fall through and let the user see the wizard
        // in its existing state.
        if (requestedOrgId && ob.org_id !== requestedOrgId) {
          try {
            ob = await completeOrganizationStep({
              body: { org_id: requestedOrgId },
            })
          } catch {
            // Fall through; the wizard will render existing state.
          }
        }
        if (!cancelled) setInitialSharedData({ onboarding: ob })
      } catch {
        if (!cancelled) setInitialSharedData({})
      }
    }
    load()
    return () => { cancelled = true }
  }, [onboardingV2, requestedOrgId])

  if (!initialSharedData) return null

  const onboarding = initialSharedData.onboarding as TOnboarding | undefined
  const initialStepIndex = onboardingV2 && onboarding?.current_step
    ? ONBOARDING_STEP_TO_INDEX[onboarding.current_step] ?? 0
    : undefined

  const wizard = (
    <OnboardingWizard
      steps={steps}
      initialSharedData={initialSharedData}
      initialStepIndex={initialStepIndex}
      onComplete={() => {
        window.location.href = '/'
      }}
    />
  )

  return (
    <ToastProvider>
      <SurfacesProvider>
        {onboardingV2 ? wizard : (
          <OnboardingJourneyProvider>
            {wizard}
          </OnboardingJourneyProvider>
        )}
      </SurfacesProvider>
    </ToastProvider>
  )
}
