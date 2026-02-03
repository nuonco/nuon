'use client'

import React, { type FC } from 'react'
import { Text } from '@/components/old/Typography'

interface ExampleApp {
  name: string
  description: string
  path: string
}

const EXAMPLE_APPS: ExampleApp[] = [
  {
    name: 'EKS Simple',
    description: 'Simple Kubernetes cluster deployment',
    path: 'eks-simple',
  },
  {
    name: 'AWS Lambda',
    description: 'Serverless functions deployment example',
    path: 'aws-lambda',
  },
  {
    name: 'AWS EC2',
    description: 'Simple EC2 deployment example',
    path: 'httpbin',
  },
  {
    name: 'Coder',
    description: 'Development environment platform',
    path: 'coder',
  },
  {
    name: 'Mattermost',
    description: 'Team collaboration platform',
    path: 'mattermost',
  },
]

interface ExampleAppsGridProps {
  selectedApp: ExampleApp | null
  onAppSelect: (app: ExampleApp) => void
  disabled?: boolean
}

export const ExampleAppsGrid: FC<ExampleAppsGridProps> = ({
  selectedApp,
  onAppSelect,
  disabled = false,
}) => {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {EXAMPLE_APPS.map((app) => {
        const isSelected = selectedApp?.path === app.path

        return (
          <div
            key={app.path}
            className={`border rounded-lg p-4 transition-colors ${
              disabled
                ? 'opacity-50 cursor-not-allowed'
                : 'cursor-pointer'
            } ${
              isSelected
                ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                : !disabled
                  ? 'border-gray-200 hover:border-gray-300 dark:border-gray-700 dark:hover:border-gray-600'
                  : 'border-gray-200 dark:border-gray-700'
            }`}
            onClick={() => !disabled && onAppSelect(app)}
          >
            <div className="flex items-start justify-between mb-2">
              <Text variant="semi-14">{app.name}</Text>
              {isSelected && (
                <div className="w-2 h-2 bg-primary-500 rounded-full" />
              )}
            </div>
            <Text className="text-gray-600 dark:text-gray-400">
              {app.description}
            </Text>
          </div>
        )
      })}
    </div>
  )
}

export { EXAMPLE_APPS }
export type { ExampleApp }
