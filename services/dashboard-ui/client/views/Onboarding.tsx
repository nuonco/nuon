import { useCallback, useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import { PageTitle } from '@/components/navigation/PageTitle'
import { OnboardingWizard } from '@/components/onboarding/OnboardingWizard'
import { CreateAppStep } from '@/components/onboarding/steps/CreateAppStep'
import { CreateInstallStep } from '@/components/onboarding/steps/CreateInstallStep'
import { CreateOrgStep } from '@/components/onboarding/steps/CreateOrgStep'
import { DownloadCliStep } from '@/components/onboarding/steps/DownloadCliStep'
import { SyncAppStep } from '@/components/onboarding/steps/SyncAppStep'
import { WelcomeStep } from '@/components/onboarding/steps/WelcomeStep'
import {
  IntroScreen,
  buildFirstRunSteps,
  resolveFirstRunOrg,
  resolveFirstRunResume,
  stepIndexFor,
  type TCloud,
  type TPath,
} from '@/components/onboarding/first-run'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import {
  useFirstRunJourney,
  type TFirstRunStep,
} from '@/hooks/use-first-run-journey'
import { trackEvent } from '@/lib/posthog-analytics'
import {
  FirstRunProvider,
  type IFirstRunContext,
} from '@/providers/first-run-provider'
import { OnboardingJourneyProvider } from '@/providers/onboarding-journey-provider'
import { OrgProvider } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import type { TAPIError } from '@/types'

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

const ONCE = {
  staleTime: Infinity,
  refetchOnWindowFocus: false,
  retry: false,
} as const

function FirstRunOnboarding() {
  const [searchParams, setSearchParams] = useSearchParams()
  // Read once, then dropped from the URL so a reload does not replay a GitHub callback.
  const [params] = useState(() => ({
    vcsConnectionId: searchParams.get('vcs-connected') ?? undefined,
    vcsError: searchParams.get('vcs-error') === '1',
    reopen: searchParams.get('reopen') === '1',
  }))
  useEffect(() => {
    if (searchParams.toString()) setSearchParams({}, { replace: true })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // The journey comes first: see ensureFirstRunJourney for why it has to exist
  // before the org does.
  const journey = useFirstRunJourney()
  const { user } = useAuth()

  const orgQuery = useQuery({
    queryKey: ['first-run-org'],
    queryFn: () => resolveFirstRunOrg({ user }),
    enabled: journey.isReady,
    ...ONCE,
  })
  const orgId = orgQuery.data?.id as string | undefined

  const resumeQuery = useQuery({
    queryKey: ['first-run-resume', orgId],
    queryFn: () =>
      resolveFirstRunResume({
        orgId: orgId!,
        journey: journey.journey,
        metadata: journey.metadata,
        forceStart:
          !!params.vcsConnectionId || params.vcsError || params.reopen,
      }),
    enabled: !!orgId && journey.isReady,
    ...ONCE,
  })
  const resume = resumeQuery.data

  const [started, setStarted] = useState<boolean>()
  const [route, setRoute] = useState<{ path: TPath; cloud: TCloud }>()
  // Bumped each time the wizard mounts, so returning from the intro starts fresh.
  const [mounts, setMounts] = useState(0)

  const path = route?.path ?? resume?.path ?? 'example'
  const cloud = route?.cloud ?? resume?.cloud ?? 'aws'
  const steps = useMemo(() => buildFirstRunSteps(path, cloud), [path, cloud])
  const isStarted = started ?? resume?.started ?? false

  const choosePath = useCallback(
    (nextPath: TPath, nextCloud: TCloud) =>
      setRoute({ path: nextPath, cloud: nextCloud }),
    []
  )
  const backToIntro = useCallback(() => setStarted(false), [])

  const onSkip = useCallback(
    async (stepId: string) => {
      try {
        await journey.skip(stepId as TFirstRunStep)
        trackEvent({ event: 'onboarding_skip', status: 'ok', user, props: { step: stepId } })
      } catch (err) {
        trackEvent({
          event: 'onboarding_skip',
          status: 'error',
          user,
          props: { step: stepId, err: (err as TAPIError)?.error },
        })
      } finally {
        window.location.assign(`/${orgId}`)
      }
    },
    [journey, orgId, user]
  )

  useEffect(() => {
    if (!orgId || (!params.vcsConnectionId && !params.vcsError)) return
    trackEvent({
      event: 'vcs_connection_create',
      status: params.vcsError ? 'error' : 'ok',
      user,
      props: { connectionId: params.vcsConnectionId, source: 'onboarding' },
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [orgId])

  const error = (orgQuery.error ??
    journey.error ??
    resumeQuery.error) as TAPIError | null

  let content
  if (error) {
    content = (
      <div className="h-screen flex bg-background">
        <ProviderError error={error} />
      </div>
    )
  } else if (!orgId || !resume) {
    content = (
      <div className="h-screen flex bg-background">
        <ProviderLoading />
      </div>
    )
  } else {
    const context: IFirstRunContext = {
      orgId,
      journey,
      vcsConnectionId: params.vcsConnectionId,
      vcsError: params.vcsError,
      choosePath,
      backToIntro,
    }
    const firstMount = mounts === 0
    content = (
      <OrgProvider orgId={orgId}>
        <FirstRunProvider value={context}>
          {isStarted ? (
            <OnboardingWizard
              key={mounts}
              onHistoryBack={started ? backToIntro : undefined}
              steps={steps}
              initialStepIndex={
                firstMount ? stepIndexFor(steps, resume.step) : 0
              }
              initialSharedData={
                firstMount
                  ? resume.sharedData
                  : { ...resume.sharedData, ...journey.metadata, path, cloud }
              }
              onSkip={onSkip}
              onComplete={() => window.location.assign(`/${orgId}`)}
            />
          ) : (
            <IntroScreen
              onStart={() => {
                if (started === false) setMounts((n) => n + 1)
                setStarted(true)
              }}
            />
          )}
        </FirstRunProvider>
      </OrgProvider>
    )
  }

  return (
    <ToastProvider>
      <PageTitle title="Onboarding" />
      <SurfacesProvider>{content}</SurfacesProvider>
    </ToastProvider>
  )
}

function ExistingOnboarding() {
  return (
    <ToastProvider>
      <PageTitle title="Onboarding" />
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

export function Onboarding() {
  const { onboardingFirstRun } = useConfig()
  if (onboardingFirstRun) return <FirstRunOnboarding />
  return <ExistingOnboarding />
}
