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
          Tell us a bit about yourself (optional).
        </Text>
        <UserProfile />
      </div>
      <div className="flex flex-col gap-4">
        <Input
          id="jobTitle"
          name="jobTitle"
          labelProps={{ labelText: 'Job Title' }}
        />
        <Input
          id="companyName"
          name="companyName"
          labelProps={{ labelText: 'Company Name' }}
        />
        <Textarea
          id="tellUsMore"
          name="tellUsMore"
          placeholder="How did you find us? What is your use case, app architecture, cloud providers?"
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
            <>{nextStepTitle ?? 'Continue'} <Icon variant="CaretRightIcon" weight="bold" /></>
          )}
        </Button>
      </div>
    </form>
  )
}
