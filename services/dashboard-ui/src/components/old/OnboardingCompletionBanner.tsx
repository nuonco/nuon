'use client'

import React, { type FC, useState, useEffect } from 'react'
import { X } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { Text } from '@/components/old/Typography'

interface OnboardingCompletionBannerProps {
  onDismiss?: () => void
  autoHideDelay?: number
}

export const OnboardingCompletionBanner: FC<
  OnboardingCompletionBannerProps
> = ({
  onDismiss,
  autoHideDelay = 10000, // 10 seconds default
}) => {
  const [isVisible, setIsVisible] = useState(true)

  const handleDismiss = () => {
    setIsVisible(false)
    // Call parent dismiss handler after animation
    setTimeout(() => {
      onDismiss?.()
    }, 300) // Match transition duration
  }

  if (!isVisible) {
    return null
  }

  return (
    <div className="relative mt-6 mx-6 bg-gradient-to-r from-green-50 to-blue-50 border border-green-200 dark:from-green-900/20 dark:to-blue-900/20 dark:border-green-800 rounded-lg p-4 shadow-sm transition-opacity duration-300">
      <div className="flex items-start justify-between">
        <div className="flex items-start space-x-3">
          <div className="flex-shrink-0">
            <div className="w-8 h-8 bg-green-100 dark:bg-green-800 rounded-full flex items-center justify-center">
              <span className="text-xl">🎉</span>
            </div>
          </div>
          <div className="flex-1">
            <Text
              variant="semi-14"
              className="text-green-800 dark:text-green-200"
            >
              Congratulations! You have successfully created your first app
              config.
            </Text>
            <Text
              variant="reg-14"
              className="mt-1 text-green-700 dark:text-green-300"
            >
              Your first install is now provisioning. You can monitor the
              progress below.
            </Text>
          </div>
        </div>
        <Button
          variant="ghost"
          onClick={handleDismiss}
          className="flex-shrink-0 p-1 !bg-transparent text-green-600 hover:text-green-800 dark:text-green-400 dark:hover:text-green-200"
        >
          <X size={20} />
        </Button>
      </div>
    </div>
  )
}
