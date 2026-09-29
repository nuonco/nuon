import { type FormEvent } from 'react'
import { useMutation } from '@tanstack/react-query'
import { updateUserJourneyStepMetadata } from '@/lib'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import { WelcomeStep } from './WelcomeStep'

export const WelcomeStepContainer = ({ onAdvance, setSharedData, nextStepTitle }: IWizardStepComponentProps) => {
  const { mutate, isPending } = useMutation({
    mutationFn: (metadata: Record<string, string>) =>
      updateUserJourneyStepMetadata({
        journeyName: 'evaluation',
        stepName: 'account_created',
        metadata,
        complete: true,
      }),
  })

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    const jobTitle = formData.get('jobTitle') as string
    const companyName = formData.get('companyName') as string
    const tellUsMore = formData.get('tellUsMore') as string

    setSharedData('jobTitle', jobTitle)
    setSharedData('companyName', companyName)
    setSharedData('tellUsMore', tellUsMore)

    mutate({ jobTitle, companyName, notes: tellUsMore })
  }

  return (
    <WelcomeStep
      isPending={isPending}
      nextStepTitle={nextStepTitle}
      onSubmit={handleSubmit}
      onAdvance={onAdvance}
    />
  )
}
