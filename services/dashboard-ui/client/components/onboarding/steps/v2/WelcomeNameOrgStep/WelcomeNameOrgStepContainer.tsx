import { useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { completeOrganizationStep, getOrg } from '@/lib'
import { useOnboardingPoll } from '@/hooks/use-onboarding-poll'
import type { TOnboarding } from '@/types'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import { WelcomeNameOrgStep } from './WelcomeNameOrgStep'

const fetchRandomName = async () => {
  const res = await fetch('/api/random-name')
  const data = await res.json()
  return data.name as string
}

export const WelcomeNameOrgStepContainer = ({
  onAdvance,
  sharedData,
  setSharedData,
  nextStepTitle,
}: IWizardStepComponentProps) => {
  const onboarding = sharedData.onboarding as TOnboarding | undefined
  const orgId = onboarding?.org_id
  const isStillProvisioning = onboarding?.status_v2?.status === 'in-progress' && !!orgId

  const [orgName, setOrgName] = useState('')
  const [waiting, setWaiting] = useState(isStillProvisioning)

  const { data: org } = useQuery({
    queryKey: ['org', orgId],
    queryFn: () => getOrg({ orgId: orgId! }),
    enabled: !!orgId,
  })

  const { mutate: generateName } = useMutation({
    mutationFn: fetchRandomName,
    onSuccess: (name) => setOrgName(name),
  })

  const { mutate: submit, isPending, error } = useMutation({
    mutationFn: () => completeOrganizationStep({ body: { name: orgName.trim() } }),
    onSuccess: (ob) => {
      setSharedData('onboarding', ob)
      if (ob.status_v2?.status === 'in-progress') {
        setWaiting(true)
      } else {
        onAdvance()
      }
    },
  })

  useOnboardingPoll({
    enabled: waiting,
    onResolved: (ob) => {
      setWaiting(false)
      setSharedData('onboarding', ob)
      if (ob.status_v2?.status === 'error') return
      onAdvance()
    },
  })

  const displayName = org?.name ?? orgName
  const displayId = org?.id ?? orgId

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!orgName.trim() || isPending || waiting) return
    submit()
  }

  return (
    <WelcomeNameOrgStep
      org={org}
      orgName={orgName}
      setOrgName={setOrgName}
      isPending={isPending}
      waiting={waiting}
      error={error}
      stepError={onboarding?.status_v2?.status === 'error' ? onboarding?.step_error : undefined}
      displayName={displayName}
      displayId={displayId}
      onSubmit={handleSubmit}
      onAdvance={onAdvance}
      onGenerateName={() => generateName()}
    />
  )
}
