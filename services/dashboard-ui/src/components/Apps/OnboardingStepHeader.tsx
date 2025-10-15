'use client'

import React, { type FC } from 'react'
import { Check } from '@phosphor-icons/react'
import { Text, Heading } from '@/components/Typography'
import type { TUserJourneyStep } from '@/types'

interface OnboardingStepHeaderProps {
  step: TUserJourneyStep
}

const STEP_INFO = {
  account_created: {
    title: 'Your Account',
  },
  org_created: {
    title: 'Set Up a Trial Organization',
  },
  cli_installed: {
    title: 'Install the Nuon CLI',
  },
  app_created: {
    title: 'Create Your First App',
  },
  app_synced: {
    title: 'Sync the App Config',
  },
  install_created: {
    title: 'Create an App Install',
  },
}

export const OnboardingStepHeader: FC<OnboardingStepHeaderProps> = ({
  step,
}) => {
  const stepInfo = STEP_INFO[step.name as keyof typeof STEP_INFO] || {
    title: step.title || step.name,
  }

  return (
    <div className="max-w-3xl mx-auto p-6 md:p-8">
      {/* Step Title */}
      <Heading className="text-2xl md:text-3xl font-bold">
        {stepInfo.title}

        {/* Completion Status */}
        {step.complete && (
          <div className="">
            <div className="inline-flex items-center gap-2 px-3 py-1 bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-200 rounded-full">
              <Check size={16} weight="bold" />
              <Text variant="reg-12" className="font-medium">
                Completed
              </Text>
            </div>
          </div>
        )}
      </Heading>
    </div>
  )
}
