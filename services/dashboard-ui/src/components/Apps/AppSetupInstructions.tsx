'use client'

import React, { type FC } from 'react'
import { Text } from '@/components/Typography'
import { CodeInline } from '@/components/Typography'
import type { ExampleApp } from './ExampleAppsGrid'

interface AppSetupInstructionsProps {
  selectedApp: ExampleApp
  appCreated?: boolean
  appId?: string
}

export const AppSetupInstructions: FC<AppSetupInstructionsProps> = ({
  selectedApp,
  appCreated = false,
  appId,
}) => {
  interface CommandStep {
    title: string
    command: string
    completed: boolean
    highlight?: boolean
  }

  const getCommands = (): CommandStep[] => {
    const baseCommands: CommandStep[] = [
      {
        title: 'Clone the repository',
        command: 'git clone https://github.com/nuonco/example-app-configs.git',
        completed: false,
      },
      {
        title: 'Navigate to the app directory',
        command: `cd example-app-configs/${selectedApp.path}`,
        completed: false,
      },
    ]

    if (appCreated) {
      // If app already created, no additional steps needed
      return baseCommands
    } else {
      // Show app creation step only
      return [
        ...baseCommands,
        {
          title: 'Create your app',
          command: 'nuon apps create -n my-app',
          completed: false,
        },
      ]
    }
  }

  const commands = getCommands()

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <Text variant="semi-18">
          {appCreated ? `Sync ${selectedApp.name} Configuration` : `Setup ${selectedApp.name}`}
        </Text>
        <Text className="text-gray-600 dark:text-gray-400">
          {appCreated
            ? 'Your app has been created. Now sync your configuration to complete the setup.'
            : `Follow these steps to create your app using the ${selectedApp.name} example configuration.`
          }
        </Text>
      </div>

      <div className="space-y-4">
        {commands.map((step, index) => (
          <div key={index} className="space-y-2">
            <Text variant="semi-14">
              {index + 1}. {step.title}
            </Text>
            <div
              className={`rounded-lg p-3 font-mono text-sm overflow-x-auto ${
                step.highlight
                  ? 'bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800'
                  : 'bg-gray-100 dark:bg-gray-800'
              }`}
            >
              <code className={`${
                step.highlight
                  ? 'text-blue-800 dark:text-blue-200'
                  : 'text-gray-800 dark:text-gray-200'
              }`}>
                {step.command}
              </code>
            </div>
          </div>
        ))}
      </div>

    </div>
  )
}