'use client'

import React, { type FC } from 'react'
import { Text } from '@/components/old/Typography'
import { CommandBlock } from '@/components/old/CommandBlock'

interface AppSyncStepContentProps {
  stepComplete: boolean
  selectedAppPath?: string
}

// Default app path for fallback
const DEFAULT_APP_PATH = 'eks-simple'
const DEFAULT_APP_NAME = 'EKS Simple'

export const AppSyncStepContent: FC<AppSyncStepContentProps> = ({
  stepComplete,
  selectedAppPath = DEFAULT_APP_PATH,
}) => {
  // Map app paths to display names
  const getAppDisplayName = (path: string): string => {
    const appNames: Record<string, string> = {
      'eks-simple': 'EKS Simple',
      'aks-simple': 'AKS Simple',
      'gke-simple': 'GKE Simple',
      // Add more mappings as needed
    }
    return appNames[path] || path
  }

  const appName = getAppDisplayName(selectedAppPath)

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
              App synced successfully!
            </Text>
          </div>
          <Text className="text-gray-600 dark:text-gray-400">
            Your {appName} configuration has been synced and builds are in progress. You can now proceed to create an install.
          </Text>
        </div>
      )}

      {/* Original Step Instructions - Always shown */}
      <div className={`space-y-3 ${stepComplete ? 'opacity-75' : ''}`}>
        <Text>
          Now sync your {appName} configuration to make it available for deployment.
        </Text>
        <Text>
          In addition to syncing your app config, this will trigger
          builds to package the component source code for deployment.
          You don&apos;t need to wait for these to finish, and can move on
          to creating an install.
        </Text>
        <CommandBlock command="nuon apps sync" />
      </div>
    </div>
  )
}
