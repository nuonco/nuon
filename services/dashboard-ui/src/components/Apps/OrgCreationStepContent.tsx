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

  // If step is already complete, show success state
  if (stepComplete) {
    return (
      <div className="space-y-3">
        <Text className="text-green-600 dark:text-green-400">
          ✅ Your trial organization has been created successfully!
        </Text>
        <Text className="text-sm text-gray-600 dark:text-gray-400">
          You can now proceed to the next steps to set up your applications and deployments.
        </Text>
      </div>
    )
  }

  // Show creating state
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

  // Show error state
  if (error) {
    return (
      <div className="space-y-3">
        <div className="flex items-center gap-2 text-red-600 dark:text-red-400">
          <Text>⚠️ Setup Failed</Text>
        </div>
        <Text className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          {error}
        </Text>
        <Button
          onClick={retry}
          variant="primary"
        >
          Try Again
        </Button>
      </div>
    )
  }

  // Default state (should trigger auto-creation via the hook)
  return (
    <div className="space-y-3">
      <Text className="text-gray-600 dark:text-gray-400">
        Welcome! We&rsquo;re preparing to create your organization.
      </Text>
      <Text className="text-sm text-gray-500 dark:text-gray-500">
        Your trial organization will be created automatically so you can get started quickly.
      </Text>
    </div>
  )
}