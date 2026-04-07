import { type FormEvent } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Textarea } from '@/components/common/form/Textarea'
import { UserProfile } from '@/components/users/UserProfile'

interface IWelcomeStep {
  isPending: boolean
  nextStepTitle?: string
  onSubmit: (e: FormEvent<HTMLFormElement>) => void
  onAdvance: () => void
}

export const WelcomeStep = ({ isPending, nextStepTitle, onSubmit, onAdvance }: IWelcomeStep) => {
  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-6">
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
