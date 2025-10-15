'use client'

import React, { type FC } from 'react'
import { Text } from '@/components/Typography'

interface CreateAppStepContentProps {
  stepComplete: boolean
}

export const CreateAccountStepContent: FC<CreateAppStepContentProps> = ({
  stepComplete,
}) => {
  return (
    <div className="space-y-6">
      <div className="space-y-3 pb-4 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 bg-green-500 rounded-full" />
          <Text
            variant="semi-14"
            className="text-green-800 dark:text-green-200"
          >
            Welcome to Nuon!
          </Text>
        </div>
        <Text>
          Your account has been created. Click &quot;Continue&quot; to create
          your trial organization and get started.
        </Text>
        <Text>
          Your account has been created. Click &quot;Continue&quot; to create
          your trial organization and get started.
        </Text>
      </div>
    </div>
  )
}

