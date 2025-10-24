'use client'

import React, { type FC } from 'react'
import { Text, Heading } from '@/components/old/Typography'
import { TAccount } from '@/types/ctl-api.types'
import { Profile } from '../Profile'
import { Card, type ICard } from '@/components/common/Card'

interface CreateAppStepContentProps {
  stepComplete: boolean
  account: TAccount
}

export const CreateAccountStepContent: FC<CreateAppStepContentProps> = ({
  stepComplete,
  account,
}) => {
  return (
    <div className="space-y-6">
      <div className="space-y-3 pb-4 border-b border-gray-200 dark:border-gray-700">
        <Heading>Welcome to Nuon!</Heading>
        <Text variant="med-14">
          Your account has been created and you are ready to get started.
        </Text>
        <Card className="max-w-80">
          <Profile />
        </Card>
        <Text variant="med-14">
          Next, we&apos;ll create an org you can use to manage apps, installs,
          and team members.
        </Text>
      </div>
    </div>
  )
}
