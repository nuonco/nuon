'use client'

import React, { type FC, useEffect } from 'react'
import { Spinner } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Text } from '@/components/Typography'
import { useAutoOrgCreation } from '@/hooks/useAutoOrgCreation'

interface OrgCreationStepContentProps {
  stepComplete: boolean
  onOrgCreated?: (orgId: string) => void
}

export const OrgCreationStepContent: FC<OrgCreationStepContentProps> = ({
  stepComplete,
  onOrgCreated,
}) => {
  const { isCreating, error, retry } = useAutoOrgCreation()

  // Early return for non-success states - these don't need the prepend pattern
  // since they're temporary states during the creation process
  if (isCreating) {
    return (
      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <Spinner className="animate-spin" size={16} />
          <Text className="text-blue-600 dark:text-blue-400">
            Setting up your workspace...
          </Text>
        </div>
        <Text className="text-sm text-gray-600 dark:text-gray-400">
          Creating your trial organization. This will only take a moment.
        </Text>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-3">
        <div className="flex items-center gap-2 text-red-600 dark:text-red-400">
          <Text>⚠️ Setup Failed</Text>
        </div>
        <Text className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          {error}
        </Text>
        <Button onClick={retry} variant="primary">
          Try Again
        </Button>
      </div>
    )
  }

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
              Your trial organization has been created successfully!
            </Text>
          </div>
          <Text className="text-gray-600 dark:text-gray-400">
            You can now proceed to the next steps to set up your applications
            and deployments.
          </Text>
        </div>
      )}

      {/* Original Step Instructions - Always shown */}
      <div className={`space-y-3 ${stepComplete ? 'opacity-75' : ''}`}>
        <Text className="text-gray-600 dark:text-gray-400">
          Welcome! We&rsquo;re preparing to create your organization.
        </Text>
        <Text className="text-sm text-gray-500 dark:text-gray-500">
          Your trial organization will be created automatically.
        </Text>
      </div>
    </div>
  )
}

