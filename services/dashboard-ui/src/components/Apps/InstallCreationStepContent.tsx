'use client'

import React, { type FC } from 'react'
import { Text } from '@/components/Typography'
import { Button } from '@/components/Button'

interface InstallCreationStepContentProps {
  stepComplete: boolean
  onClose: () => void
  installId?: string
}

export const InstallCreationStepContent: FC<InstallCreationStepContentProps> = ({
  stepComplete,
  onClose,
  installId,
}) => {
  if (stepComplete) {
    return (
      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 bg-green-500 rounded-full" />
          <Text variant="semi-14" className="text-green-800 dark:text-green-200">
            Install created successfully!
          </Text>
        </div>
        <Text className="text-gray-600 dark:text-gray-400">
          Great! Your first install has been created and you&apos;re ready to deploy your app.
        </Text>
        {installId && (
          <div className="bg-gray-50 dark:bg-gray-800 p-3 rounded-lg">
            <Text variant="reg-12" className="text-gray-600 dark:text-gray-400">
              Install ID: <code className="font-mono text-gray-800 dark:text-gray-200">{installId}</code>
            </Text>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="space-y-3">
        <Text className="text-gray-800 dark:text-gray-200">
          You are now ready to create your first install. Close this modal and click on the &quot;New Install&quot; button.
        </Text>
      </div>

      <div className="flex justify-start">
        <Button
          onClick={onClose}
          variant="primary"
        >
          Got it
        </Button>
      </div>
    </div>
  )
}