'use client'

import React, { type FC, useState } from 'react'
import { usePathname } from 'next/navigation'
import { Text } from '@/components/old/Typography'
import { Button } from '@/components/old/Button'
import { ExampleAppsGrid, EXAMPLE_APPS, type ExampleApp } from './ExampleAppsGrid'
import { createAppFromTemplate } from '@/actions/apps/create-app-from-template'
import { useAccount } from '@/hooks/use-account'
import { useServerAction } from '@/hooks/use-server-action'

interface CreateAppStepContentProps {
  stepComplete: boolean
  appId?: string
  orgId?: string
}

export const CreateAppStepContent: FC<CreateAppStepContentProps> = ({
  stepComplete,
  appId,
  orgId,
}) => {
  const pathname = usePathname()
  const [selectedApp, setSelectedApp] = useState<ExampleApp>(EXAMPLE_APPS[0])
  const [verifying, setVerifying] = useState(false)
  const [confirmedError, setConfirmedError] = useState(false)
  const { refreshAccount } = useAccount()

  const { data, error, isLoading, execute } = useServerAction({
    action: createAppFromTemplate,
  })

  const handleAppCreate = async (app: ExampleApp) => {
    if (!orgId) return

    setVerifying(false)
    setConfirmedError(false)

    const result = await execute({
      body: { template: app.path },
      orgId,
      path: pathname,
    })

    // If we got data back, success is immediate
    if (result?.data) {
      return
    }

    // If there was an error (e.g., timeout), verify with account polling
    if (result?.error) {
      setVerifying(true)

      // Poll account multiple times to confirm if app was actually created
      for (let i = 0; i < 5; i++) {
        await new Promise((resolve) => setTimeout(resolve, 2000))
        await refreshAccount()
        // If stepComplete becomes true during polling, we're done (success will show via props)
        // We can't check stepComplete here directly, but the component will re-render
      }

      // After polling, if we're still in verifying state (not success), show error
      setVerifying(false)
      setConfirmedError(true)
    }
  }

  // Success determined by props (stepComplete) or server action data
  const isSuccess = stepComplete || !!data

  // Only show error if we've confirmed failure after verification
  // AND we're not in a success state (props take precedence)
  const showError = confirmedError && !isSuccess && !verifying

  const isDisabled = isSuccess || isLoading || verifying || !orgId

  return (
    <div className="space-y-6">
      {isSuccess && (
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
            step to create your first install.
          </Text>
          {(appId || data?.id) && (
            <div className="bg-gray-50 dark:bg-gray-800 p-3 rounded-lg">
              <Text
                variant="reg-12"
                className="text-gray-600 dark:text-gray-400"
              >
                App ID:{' '}
                <code className="font-mono text-gray-800 dark:text-gray-200">
                  {appId || data?.id}
                </code>
              </Text>
            </div>
          )}
        </div>
      )}

      {verifying && !isSuccess && (
        <div className="space-y-3 pb-4 border-b">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-yellow-500 rounded-full animate-pulse" />
            <Text
              variant="semi-14"
              className="text-yellow-800 dark:text-yellow-200"
            >
              Verifying app creation...
            </Text>
          </div>
          <Text className="text-gray-600 dark:text-gray-400">
            Please wait while we confirm your app was created.
          </Text>
        </div>
      )}

      {showError && (
        <div className="space-y-3 pb-4 border-b">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-red-500 rounded-full" />
            <Text
              variant="semi-14"
              className="text-red-800 dark:text-red-200"
            >
              Failed to create app
            </Text>
          </div>
          <Text className="text-gray-600 dark:text-gray-400">
            {error?.error || error?.description || 'An error occurred. Please try again.'}
          </Text>
        </div>
      )}

      <div className={`space-y-6 ${isSuccess ? 'opacity-75' : ''}`}>
        <div className="space-y-3">
          <Text variant="semi-14">Choose your example app</Text>
          <Text variant="reg-12" className="text-gray-600 dark:text-gray-400">
            Select an example app to get started. We&apos;ll create and
            configure it automatically.
          </Text>
          <ExampleAppsGrid
            selectedApp={selectedApp}
            onAppSelect={setSelectedApp}
            disabled={isDisabled}
          />
          <div className="flex items-center justify-between p-3 bg-primary-50 dark:bg-primary-900/20 border border-primary-200 dark:border-primary-800 rounded-lg">
            <Text variant="reg-12" className="text-primary-800 dark:text-primary-200">
              Selected: <strong>{selectedApp.name}</strong>
            </Text>
            <Button
              variant="primary"
              onClick={() => handleAppCreate(selectedApp)}
              disabled={isDisabled}
            >
              {isLoading ? 'Creating...' : 'Create App'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
