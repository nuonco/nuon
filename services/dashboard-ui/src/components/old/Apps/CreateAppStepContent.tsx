'use client'

import React, { type FC, useState } from 'react'
import { usePathname } from 'next/navigation'
import { Text } from '@/components/old/Typography'
import { ExampleAppsGrid, type ExampleApp } from './ExampleAppsGrid'
import { createAppFromTemplate } from '@/actions/apps/create-app-from-template'
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
  const [creatingApp, setCreatingApp] = useState<string | null>(null)

  const { data, error, isLoading, execute } = useServerAction({
    action: createAppFromTemplate,
  })

  const handleAppCreate = async (app: ExampleApp) => {
    if (!orgId) return

    setCreatingApp(app.path)
    await execute({
      body: { template: app.path },
      orgId,
      path: pathname,
    })
    setCreatingApp(null)
  }

  return (
    <div className="space-y-6">
      {(stepComplete || data) && (
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

      {error && !data && (
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

      <div className={`space-y-6 ${stepComplete || data ? 'opacity-75' : ''}`}>
        <div className="space-y-3">
          <Text variant="semi-14">Choose your example app</Text>
          <Text variant="reg-12" className="text-gray-600 dark:text-gray-400">
            Select an example app to get started. We&apos;ll create and
            configure it automatically.
          </Text>
          <ExampleAppsGrid
            onAppCreate={handleAppCreate}
            creatingApp={creatingApp}
            disabled={stepComplete || !!data || isLoading || !orgId}
          />
        </div>
      </div>
    </div>
  )
}
