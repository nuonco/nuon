'use client'

import React, { type FC } from 'react'
import { Text } from '@/components/Typography'
import { CodeInline } from '@/components/Typography'
import { CommandBlock } from '@/components/CommandBlock'
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
      {
        title: 'Create your app',
        command: `nuon apps create -n ${selectedApp.path}`,
        completed: false,
      },
    ]
    return baseCommands
  }

  const commands = getCommands()

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <Text variant="semi-18">
          {appCreated
            ? `Sync ${selectedApp.name} Configuration`
            : `Setup ${selectedApp.name}`}
        </Text>
        <Text className="text-gray-600 dark:text-gray-400">
          {appCreated
            ? 'Your app has been created. Now sync your configuration to complete the setup.'
            : `Follow these steps to create your app using the ${selectedApp.name} example configuration.`}
        </Text>
      </div>

      <div className="space-y-4">
        {commands.map((step, index) => (
          <CommandBlock
            key={index}
            command={step.command}
            title={`${index + 1}. ${step.title}`}
          />
        ))}
      </div>
    </div>
  )
}
