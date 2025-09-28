'use client'

import React, { type FC } from 'react'
import { Text } from '@/components/Typography'
import { CommandBlock } from '@/components/CommandBlock'

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
            Great! The Nuon CLI is now installed and ready to use. You can proceed
            to the next step.
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
        </div>

        <div className="space-y-4">
          <Text variant="semi-14">1. Install the CLI</Text>
          {installMethods.map((method, index) => (
            <CommandBlock
              key={index}
              command={method.command}
              title={method.title}
              description={method.description}
            />
          ))}
        </div>

        {/* Login Step */}
        <CommandBlock command="nuon login" title="2. Log in to your account" />

        {/* Select Org Step */}
        <CommandBlock command="nuon orgs select" title="3. Select your org" />
      </div>
    </div>
  )
}
