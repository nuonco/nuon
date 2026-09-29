import { useState } from 'react'
import { keepPreviousData, useMutation, useQuery } from '@tanstack/react-query'
import { createOrg, getOrg } from '@/lib'
import { trackEvent } from '@/lib/posthog-analytics'
import { useAuth } from '@/hooks/use-auth'
import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import type { TOrg } from '@/types'
import { CreateOrgStep, CompletedOrgCard } from './CreateOrgStep'

export const CreateOrgStepContainer = ({
  onAdvance,
  nextStepTitle,
  setSharedData,
}: IWizardStepComponentProps) => {
  const [createdOrg, setCreatedOrg] = useState<TOrg | null>(null)
  const { user } = useAuth()
  const { isStepComplete, getStepMetadata } = useOnboardingJourney()

  const orgCreated = isStepComplete('org_created')
  const journeyOrgId = getStepMetadata('org_created', 'org_id') as
    | string
    | undefined

  const { mutate, isPending, error } = useMutation({
    mutationFn: (name: string) =>
      createOrg({ body: { name, use_sandbox_mode: false, tags: ['Trial'] } }),
    onSuccess: (org) => {
      trackEvent({
        event: 'org_create',
        status: 'ok',
        user,
        props: { orgId: org.id },
      })
      setCreatedOrg(org)
      setSharedData('orgId', org.id)
    },
    onError: (err: any) => {
      trackEvent({
        event: 'org_create',
        status: 'error',
        user,
        props: { err: err?.error },
      })
    },
  })

  const { mutateAsync: generateName } = useMutation({
    mutationFn: async () => {
      const res = await fetch('/api/random-name')
      const data = await res.json()
      return data.name as string
    },
  })

  if (orgCreated && journeyOrgId && !createdOrg) {
    return (
      <CompletedOrgCardContainer
        orgId={journeyOrgId}
        onAdvance={onAdvance}
        nextStepTitle={nextStepTitle}
      />
    )
  }

  return (
    <CreateOrgStep
      onAdvance={onAdvance}
      nextStepTitle={nextStepTitle}
      createdOrg={createdOrg}
      isPending={isPending}
      error={error}
      onCreateOrg={(name) => mutate(name)}
      onGenerateName={() => generateName()}
    />
  )
}

function CompletedOrgCardContainer({
  orgId,
  onAdvance,
  nextStepTitle,
}: {
  orgId: string
  onAdvance: IWizardStepComponentProps['onAdvance']
  nextStepTitle: IWizardStepComponentProps['nextStepTitle']
}) {
  const { data: org, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['onboarding-org', orgId],
    queryFn: () => getOrg({ orgId }),
    refetchInterval: 10000,
  })

  return (
    <CompletedOrgCard
      org={org}
      orgId={orgId}
      isLoading={isLoading}
      onAdvance={onAdvance}
      nextStepTitle={nextStepTitle}
    />
  )
}
