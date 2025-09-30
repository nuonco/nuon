'use client'

import React, { type FC } from 'react'
import { Text } from '@/components/Typography'
import { CommandBlock } from '@/components/CommandBlock'
import { ClickToCopyButton } from '@/components/ClickToCopy'

interface CLIInstallStepContentProps {
  stepComplete: boolean
}

export const CLIInstallStepContent: FC<CLIInstallStepContentProps> = ({
  stepComplete,
}) => {
  const installMethods = [
    {
      title: 'macOS (Homebrew)',
      command: 'brew install nuonco/tap/nuon',
      description: 'Recommended for macOS users',
    },
  ]

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
              Nuon CLI installed successfully!
            </Text>
          </div>
          <Text className="text-gray-600 dark:text-gray-400">
            Great! The Nuon CLI is now installed and ready to use. You can
            proceed to the next step.
          </Text>
        </div>
      )}

      {/* Original Step Instructions - Always shown */}
      <div className={`space-y-6 ${stepComplete ? 'opacity-75' : ''}`}>
        <div className="space-y-2">
          <Text className="text-gray-600 dark:text-gray-400">
            The Nuon CLI is required to create and manage your applications.
            Choose your preferred installation method and login to get started.
          </Text>
          <Text className="text-gray-600 dark:text-gray-400">
            If you are not using homebrew, see{' '}
            <a
              href="https://docs.nuon.co/cli"
              target="_"
              className="text-active font-medium"
            >
              our CLI docs
            </a>{' '}
            for other installation options.
          </Text>
        </div>

        <div className="space-y-4">
          <Text variant="semi-14">1. Install the CLI</Text>
          <div className="space-y-2">
            <div className="relative rounded-lg p-3 font-mono text-sm bg-gray-100 dark:bg-gray-800">
              <div className="flex items-center justify-between gap-2">
                <code className="text-gray-800 dark:text-gray-200 flex-1 min-w-0">
                  brew install nuonco/tap/nuon
                </code>
                <ClickToCopyButton
                  textToCopy="brew install nuonco/tap/nuon"
                  className="opacity-70 hover:opacity-100 flex-shrink-0"
                />
              </div>
            </div>
          </div>
        </div>

        {/* Login Step */}
        <CommandBlock command="nuon login" title="2. Log in to your account" />

        {/* Select Org Step */}
        <CommandBlock command="nuon orgs select" title="3. Select your org" />
      </div>
    </div>
  )
}
