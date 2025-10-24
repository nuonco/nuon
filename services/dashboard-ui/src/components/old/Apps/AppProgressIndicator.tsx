'use client'

import React, { type FC } from 'react'
import { Text } from '@/components/old/Typography'

const CheckCircleIcon: FC<{ className?: string }> = ({ className }) => (
  <svg
    className={className}
    fill="currentColor"
    viewBox="0 0 20 20"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path
      fillRule="evenodd"
      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
      clipRule="evenodd"
    />
  </svg>
)

const ClockIcon: FC<{ className?: string }> = ({ className }) => (
  <svg
    className={className}
    fill="currentColor"
    viewBox="0 0 20 20"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path
      fillRule="evenodd"
      d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z"
      clipRule="evenodd"
    />
  </svg>
)

interface AppProgressIndicatorProps {
  appCreated: boolean
  appSynced: boolean
  className?: string
}

export const AppProgressIndicator: FC<AppProgressIndicatorProps> = ({
  appCreated,
  appSynced,
  className = '',
}) => {
  const steps = [
    {
      name: 'Create App',
      description: 'Create your app configuration',
      completed: appCreated,
      current: !appCreated,
    },
    {
      name: 'Sync Config',
      description: 'Run CLI command to sync configuration',
      completed: appSynced,
      current: appCreated && !appSynced,
    },
  ]

  return (
    <div className={`space-y-4 ${className}`}>
      <Text variant="semi-14" className="text-gray-900 dark:text-gray-100">
        Setup Progress
      </Text>

      <div className="space-y-3">
        {steps.map((step, stepIdx) => (
          <div key={step.name} className="flex items-center">
            <div className="flex-shrink-0">
              {step.completed ? (
                <CheckCircleIcon
                  className="h-5 w-5 text-green-500"
                  aria-hidden="true"
                />
              ) : step.current ? (
                <ClockIcon
                  className="h-5 w-5 text-blue-500 animate-pulse"
                  aria-hidden="true"
                />
              ) : (
                <div className="h-5 w-5 rounded-full border-2 border-gray-300 dark:border-gray-600" />
              )}
            </div>

            <div className="ml-3 min-w-0 flex-1">
              <Text
                variant="semi-14"
                className={`${
                  step.completed
                    ? 'text-green-700 dark:text-green-400'
                    : step.current
                      ? 'text-blue-700 dark:text-blue-400'
                      : 'text-gray-500 dark:text-gray-400'
                }`}
              >
                {step.name}
              </Text>
              <Text
                className={`text-sm ${
                  step.completed
                    ? 'text-green-600 dark:text-green-500'
                    : step.current
                      ? 'text-blue-600 dark:text-blue-500'
                      : 'text-gray-400 dark:text-gray-500'
                }`}
              >
                {step.description}
              </Text>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
