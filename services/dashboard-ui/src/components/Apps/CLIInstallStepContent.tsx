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
          </Text>
        </div>

        <CommandBlock
          command="brew install nuonco/tap/nuon"
          title="1. Install the CLI"
          description={
            <span>
              If you are not using homebrew, see{' '}
              <a
                href="https://docs.nuon.co/cli"
                target="_"
                className="text-active font-medium"
              >
                our CLI docs
              </a>
            </span>
          }
        />

        <CommandBlock
          command="nuon auth login"
          title="2. Log in to your account"
          description='When prompted, select "Nuon Cloud"'
        />
      </div>
    </div>
  )
}
