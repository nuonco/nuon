import { type FormEvent } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Textarea } from '@/components/common/form/Textarea'
import { UserProfile } from '@/components/users/UserProfile'
import { updateUserJourneyStepMetadata } from '@/lib'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

export const WelcomeStep = ({ onAdvance, setSharedData, nextStepTitle }: IWizardStepComponentProps) => {
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
    <form onSubmit={handleSubmit} className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <Text variant="body" theme="neutral">
          Before we get started, tell us a bit about yourself.
        </Text>
        <UserProfile />
      </div>
      <div className="flex flex-col gap-4">
        <Input
          id="jobTitle"
          name="jobTitle"
          placeholder="e.g. Platform Engineer"
          labelProps={{ labelText: 'Job Title' }}
        />
        <Input
          id="companyName"
          name="companyName"
          placeholder="e.g. Acme Corp"
          labelProps={{ labelText: 'Company Name' }}
        />
        <Textarea
          id="tellUsMore"
          name="tellUsMore"
          placeholder="Tell us about your use case..."
          labelProps={{ labelText: 'Tell us more' }}
          rows={3}
        />
      </div>
      <div className="flex self-end">
        <Button type="submit" variant="primary" disabled={isPending} onClick={onAdvance}>
          {isPending ? (
            <span className="flex items-center gap-2">
              <Icon variant="Loading" />
              Saving...
            </span>
          ) : (
            <>{nextStepTitle ?? 'Continue'} <Icon variant="CaretRight" weight="bold" /></>
          )}
        </Button>
      </div>
    </form>
  )
}
