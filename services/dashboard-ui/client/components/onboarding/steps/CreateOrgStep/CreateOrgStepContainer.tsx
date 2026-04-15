import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { createOrg, getOrg } from '@/lib'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import type { TOrg } from '@/types'
import { CreateOrgStep, CompletedOrgCard } from './CreateOrgStep'

export const CreateOrgStepContainer = ({
  onAdvance,
  nextStepTitle,
  setSharedData,
  sharedData,
}: IWizardStepComponentProps) => {
  const [createdOrg, setCreatedOrg] = useState<TOrg | null>(null)
  const [orgName, setOrgName] = useState('')
  const { user } = useAuth()
  const { isByoc, sfTrialEndpoint } = useConfig()
  const { isStepComplete, getStepMetadata } = useOnboardingJourney()

  const orgCreated = isStepComplete('org_created')
  const journeyOrgId = getStepMetadata('org_created', 'org_id') as
    | string
    | undefined

  const { mutate, isPending, error } = useMutation({
    mutationFn: (name: string) =>
      createOrg({ body: { name, use_sandbox_mode: false, tags: ['Trial'] } }),
    onSuccess: (org) => {
      setCreatedOrg(org)
      setSharedData('orgId', org.id)

      if (!isByoc && sfTrialEndpoint) {
        const nameParts = (user?.name ?? '').split(' ')
        const firstName = nameParts[0] ?? ''
        const lastName = nameParts.slice(1).join(' ') || 'ULN'
        fetch(sfTrialEndpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            firstName,
            lastName,
            email: user?.email,
            companyName: `${sharedData.companyName ?? ''} | ${org.name}`,
            jobTitle: sharedData.jobTitle,
            notes: sharedData.tellUsMore,
            subject: 'trial-signup',
          }),
        }).catch(() => {})
      }
    },
  })

  const { mutate: generateName } = useMutation({
    mutationFn: async () => {
      const res = await fetch('/api/random-name')
      const data = await res.json()
      return data.name as string
    },
    onSuccess: (name) => setOrgName(name),
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
      orgName={orgName}
      onOrgNameChange={setOrgName}
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
