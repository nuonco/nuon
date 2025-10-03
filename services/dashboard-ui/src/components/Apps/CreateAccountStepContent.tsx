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
      {/* Success Message - Shown when step is complete */}
      {stepComplete && (
        <div className="space-y-3 pb-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full" />
            <Text
              variant="semi-14"
              className="text-green-800 dark:text-green-200"
            >
              Your account has been created successfully!
            </Text>
          </div>
          <Text className="text-gray-600 dark:text-gray-400">
            You&rsquo;re now ready to set up your organization and start
            deploying applications.
          </Text>
        </div>
      )}

      {/* Original Step Instructions - Always shown */}
      <div className={`space-y-3 ${stepComplete ? 'opacity-75' : ''}`}>
        <Text className="text-gray-600 dark:text-gray-400">
          Welcome to Nuon! Your account creation is the first step in setting up
          your deployment platform.
        </Text>
        <Text className="text-sm text-gray-500 dark:text-gray-500">
          With your account created, you can now proceed to create an
          organization and start managing your applications.
        </Text>
      </div>
    </div>
  )
}
