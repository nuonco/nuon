'use client'

import React, { type FC } from 'react'
import { Text } from '@/components/Typography'

interface CLIInstallStepContentProps {
  stepComplete: boolean
}

export const CLIInstallStepContent: FC<CLIInstallStepContentProps> = ({
  stepComplete,
}) => {
  if (stepComplete) {
    return (
      <div className="space-y-3">
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
          Great! The Nuon CLI is now installed and ready to use. You can proceed
          to the next step.
        </Text>
      </div>
    )
  }

  const installMethods = [
    {
      title: 'macOS (Homebrew)',
      command: 'brew install nuonco/nuon',
      description: 'Recommended for macOS users',
    },
    {
      title: 'Manual Installation',
      command: 'curl -sSL https://install.nuon.co | bash',
      description: 'Works on macOS, Linux, and Windows (WSL)',
    },
    {
      title: 'Direct Download',
      command: 'Download from https://github.com/nuonco/nuon-cli/releases',
      description: 'Download the binary directly for your platform',
    },
  ]

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <Text className="text-gray-600 dark:text-gray-400">
          The Nuon CLI is required to create and manage your applications.
          Choose your preferred installation method and login to get started.
        </Text>
      </div>

      <div className="space-y-4">
        <Text variant="semi-14">1. Install the CLI</Text>
        {installMethods.map((method, index) => (
          <div key={index} className="space-y-2">
            <div className="flex items-center justify-end">
              <Text
                variant="reg-12"
                className="text-gray-500 dark:text-gray-400"
              >
                {method.description}
              </Text>
            </div>
            <div className="rounded-lg p-3 font-mono text-sm overflow-x-auto bg-gray-100 dark:bg-gray-800">
              <code className="text-gray-800 dark:text-gray-200">
                {method.command}
              </code>
            </div>
          </div>
        ))}
      </div>

      {/* Login Step */}
      <div className="space-y-2">
        <Text variant="semi-14">2. Log in to your account</Text>
        <div className="rounded-lg p-3 font-mono text-sm overflow-x-auto bg-gray-100 dark:bg-gray-800">
          <code className="text-gray-800 dark:text-gray-200">
            nuon auth login
          </code>
        </div>
      </div>

      {/* Select Org Step */}
      <div className="space-y-2">
        <Text variant="semi-14">3. Select your org</Text>
        <div className="rounded-lg p-3 font-mono text-sm overflow-x-auto bg-gray-100 dark:bg-gray-800">
          <code className="text-gray-800 dark:text-gray-200">
            nuon orgs select
          </code>
        </div>
      </div>
    </div>
  )
}

