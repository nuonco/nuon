'use client'

import React, { type FC, useState } from 'react'
import { Text } from '@/components/old/Typography'
import { ExampleAppsGrid, type ExampleApp } from './ExampleAppsGrid'
import { AppSetupInstructions } from './AppSetupInstructions'

interface CreateAppStepContentProps {
  stepComplete: boolean
  appId?: string
}

// Default app for simplified onboarding
const DEFAULT_APP: ExampleApp = {
  name: 'EKS Simple',
  description: 'Simple Kubernetes cluster deployment',
  path: 'eks-simple',
}

export const CreateAppStepContent: FC<CreateAppStepContentProps> = ({
  stepComplete,
  appId,
}) => {
  const [selectedApp, setSelectedApp] = useState<ExampleApp>(DEFAULT_APP)

  return (
    <div className="space-y-6">
      {/* Success Message - Shown when step is complete */}
      {stepComplete && (
        <div className="space-y-3 pb-4 border-b">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full" />
            <Text
              variant="semi-14"
              className="text-green-800 dark:text-green-200"
            >
              App created successfully!
            </Text>
          </div>
          <Text className="text-gray-600 dark:text-gray-400">
            Your app is now configured and ready. You can proceed to the next
            step to sync your app configuration.
          </Text>
          {appId && (
            <div className="bg-gray-50 dark:bg-gray-800 p-3 rounded-lg">
              <Text
                variant="reg-12"
                className="text-gray-600 dark:text-gray-400"
              >
                App ID:{' '}
                <code className="font-mono text-gray-800 dark:text-gray-200">
                  {appId}
                </code>
              </Text>
            </div>
          )}
        </div>
      )}

      {/* Original Step Instructions - Always shown */}
      <div className={`space-y-6 ${stepComplete ? 'opacity-75' : ''}`}>
        {/* Example Apps Section */}
        <div className="space-y-3">
          <Text variant="semi-14">Choose your example app</Text>
          <Text variant="reg-12" className="text-gray-600 dark:text-gray-400">
            Start with a curated example to learn Nuon patterns and get up and
            running quickly.
          </Text>
          <ExampleAppsGrid
            selectedApp={selectedApp}
            onAppSelect={setSelectedApp}
          />
          <div className="flex items-center justify-between p-3 bg-primary-50 dark:bg-primary-900/20 border border-primary-200 dark:border-primary-800 rounded-lg">
            <Text
              variant="reg-12"
              className="text-primary-800 dark:text-primary-200"
            >
              Selected: <strong>{selectedApp.name}</strong>
            </Text>
          </div>
        </div>

        {/* Setup Instructions Section */}
        <div className="space-y-3">
          <div className="border-t border-gray-200 dark:border-gray-700 pt-6">
            <AppSetupInstructions
              selectedApp={selectedApp}
              appCreated={stepComplete}
              appId={appId}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
