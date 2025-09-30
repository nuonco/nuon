'use client'

import React, { type FC } from 'react'
import { useRouter } from 'next/navigation'
import { Text } from '@/components/Typography'
import { Button } from '@/components/Button'

interface InstallCreationStepContentProps {
  stepComplete: boolean
  onClose: () => void
  installId?: string
  appId?: string
  orgId?: string
  onNavigateToInstall?: (appId: string, orgId: string) => void
}

export const InstallCreationStepContent: FC<
  InstallCreationStepContentProps
> = ({
  stepComplete,
  onClose,
  installId,
  appId,
  orgId,
  onNavigateToInstall,
}) => {
  const router = useRouter()
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
              Install created successfully!
            </Text>
          </div>
          {installId && (
            <div className="bg-gray-50 dark:bg-gray-800 p-3 rounded-lg">
              <Text
                variant="reg-12"
                className="text-gray-600 dark:text-gray-400"
              >
                Install ID:{' '}
                <code className="font-mono text-gray-800 dark:text-gray-200">
                  {installId}
                </code>
              </Text>
            </div>
          )}
        </div>
      )}

      {/* Original Step Instructions - Always shown */}
      <div className={`space-y-4 ${stepComplete ? 'opacity-75' : ''}`}>
        <div className="space-y-3">
          <Text className="text-gray-800 dark:text-gray-200">
            You are almost ready to create an install.
          </Text>
          <Text className="text-gray-800 dark:text-gray-200">
            Before creating the install, log into the AWS account you want to
            install into. Make sure you have the required permissions to create
            resources, and have not reached any quota limits.
          </Text>
          <Text className="text-gray-800 dark:text-gray-200">
            Once your account is ready, click the &ldquo;Create Install&rdquo;
            button.
          </Text>
        </div>

        <div className="flex justify-start">
          <Button
            onClick={() => {
              // If we have app and org info and navigation callback, use coordinated flow
              if (appId && orgId && onNavigateToInstall) {
                onNavigateToInstall(appId, orgId)
              } else if (appId && orgId) {
                // Fallback to direct navigation (old behavior)
                router.push(`/${orgId}/apps/${appId}?createInstall=true`)
              } else {
                // Fallback to original behavior if navigation info is missing
                onClose()
              }
            }}
            variant="primary"
          >
            Create install
          </Button>
        </div>
      </div>
    </div>
  )
}
