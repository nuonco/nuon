'use client'

import React, { type FC } from 'react'
import { Button } from '@/components/old/Button'
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
  onAppCreate: (app: ExampleApp) => void
  creatingApp: string | null
  disabled?: boolean
}

export const ExampleAppsGrid: FC<ExampleAppsGridProps> = ({
  onAppCreate,
  creatingApp,
  disabled = false,
}) => {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {EXAMPLE_APPS.map((app) => {
        const isCreating = creatingApp === app.path
        const isDisabled = disabled || (creatingApp !== null && !isCreating)

        return (
          <div
            key={app.path}
            className={`border rounded-lg p-4 transition-colors ${
              isCreating
                ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                : 'border-gray-200 dark:border-gray-700'
            }`}
          >
            <div className="flex items-start justify-between mb-2">
              <Text variant="semi-14">{app.name}</Text>
              {isCreating && (
                <div className="w-2 h-2 bg-primary-500 rounded-full animate-pulse" />
              )}
            </div>
            <Text className="text-gray-600 dark:text-gray-400 mb-4">
              {app.description}
            </Text>
            <Button
              variant="primary"
              size="sm"
              onClick={(e) => {
                e.stopPropagation()
                onAppCreate(app)
              }}
              disabled={isDisabled}
              className="w-full"
            >
              {isCreating ? 'Creating...' : 'Select'}
            </Button>
          </div>
        )
      })}
    </div>
  )
}

export { EXAMPLE_APPS }
export type { ExampleApp }
