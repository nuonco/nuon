'use client'

import React, { type FC } from 'react'
import { Text, Heading } from '@/components/old/Typography'
import { TAccount } from '@/types/ctl-api.types'
import { Profile } from '../Profile'
import { Card } from '@/components/common/Card'
import { Input } from '@/components/common/form/Input'
import { Textarea } from '@/components/common/form/Textarea'
import { useUserJourney } from '@/hooks/use-user-journey'
import { updateUserJourneyStepMetadata } from '@/actions/accounts/update-user-journey-step-metadata'

interface CreateAppStepContentProps {
  stepComplete: boolean
  account: TAccount
  setSFData: (sfData: any) => void
  initialValues?: {
    jobTitle?: string
    companyName?: string
    notes?: string
  }
}

export const CreateAccountStepContent: FC<CreateAppStepContentProps> = ({
  stepComplete,
  account,
  setSFData,
  initialValues,
}) => {
  const { isBYOC } = useUserJourney()
  return (
    <div className="space-y-6">
      <div className="space-y-3 pb-4 border-b border-gray-200 dark:border-gray-700">
        <Heading>Welcome to Nuon!</Heading>
        <Text variant="med-14">
          Your account has been created and you are ready to get started.
        </Text>
        <div className="flex flex-col gap-6">
          <Card className="max-w-80">
            <Profile />
          </Card>
          <Card>
            <form
              id="sf-form"
              className="flex flex-col gap-4"
              onSubmit={async (e: React.FormEvent<HTMLFormElement>) => {
                e?.preventDefault()
                const formData = new FormData(e.currentTarget)
                const formObject = Object.fromEntries(formData.entries())
                setSFData?.(formObject as Record<string, string>)

                await updateUserJourneyStepMetadata({
                  journeyName: 'evaluation',
                  stepName: 'account_created',
                  metadata: {
                    jobTitle: (formObject.jobTitle as string) || '',
                    companyName: (formObject.companyName as string) || '',
                    notes: (formObject.notes as string) || '',
                  },
                  complete: true,
                })
              }}
            >
              {isBYOC ? (
                <Input
                  className="font-sans"
                  labelProps={{
                    labelText: 'Organization name',
                  }}
                  name="name"
                  defaultValue={`${account?.email}-trial`}
                  placeholder="Organization name"
                />
              ) : null}
              <Input
                className="font-sans"
                labelProps={{
                  labelText: 'Job title',
                }}
                name="jobTitle"
                placeholder="Job title"
                defaultValue={initialValues?.jobTitle}
              />
              <Input
                className="font-sans"
                labelProps={{
                  labelText: 'Company name',
                }}
                name="companyName"
                placeholder="Company name"
                defaultValue={initialValues?.companyName}
              />

              <Textarea
                className="font-sans"
                labelProps={{
                  labelText: 'Tell us more',
                }}
                name="notes"
                placeholder="To help us improve Nuon, please tell us about your use case, your app's architecture, and your cloud providers."
                rows={4}
                defaultValue={initialValues?.notes}
              />
            </form>
          </Card>
        </div>
        <Text variant="med-14">
          Next, we&apos;ll create an org you can use to manage apps, installs,
          and team members.
        </Text>
      </div>
    </div>
  )
}
