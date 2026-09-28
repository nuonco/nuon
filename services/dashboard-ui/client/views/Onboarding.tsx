import { useCallback, useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import { PageTitle } from '@/components/navigation/PageTitle'
import { OnboardingWizard } from '@/components/onboarding/OnboardingWizard'
import {
  IntroScreen,
  buildFirstRunSteps,
  resolveFirstRunOrg,
  resolveFirstRunResume,
  stepIndexFor,
  type TCloud,
  type TPath,
} from '@/components/onboarding/first-run'
import {
  useFirstRunJourney,
  type TFirstRunStep,
} from '@/hooks/use-first-run-journey'
import {
  FirstRunProvider,
  type IFirstRunContext,
} from '@/providers/first-run-provider'
import { OrgProvider } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import type { TAPIError } from '@/types'

const ONCE = {
  staleTime: Infinity,
  refetchOnWindowFocus: false,
  retry: false,
} as const

export function Onboarding() {
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

  const orgQuery = useQuery({
    queryKey: ['first-run-org'],
    queryFn: resolveFirstRunOrg,
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
      } finally {
        window.location.assign(`/${orgId}`)
      }
    },
    [journey, orgId]
  )

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
