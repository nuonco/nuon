// TODO: this should be moved to the components/old dir

'use client'

import React, { type FC, useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { OnboardingCompletionBanner } from '@/components/old/OnboardingCompletionBanner'
import { useConfetti } from '@/hooks/use-confetti'

interface OnboardingCelebrationWrapperProps {
  children: React.ReactNode
}

export const OnboardingCelebrationWrapper: FC<OnboardingCelebrationWrapperProps> = ({
  children,
}) => {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [showCelebration, setShowCelebration] = useState(false)
  const { fireCelebrationConfetti } = useConfetti()

  useEffect(() => {
    // Check for onboarding completion parameter
    if (searchParams.get('onboardingComplete') === 'true') {
      setShowCelebration(true)

      // Trigger confetti animation after a short delay to let banner appear
      const confettiTimeout = setTimeout(() => {
        fireCelebrationConfetti()
      }, 300)

      // Clean up URL parameter to prevent issues with refresh/back button
      const url = new URL(window.location.href)
      url.searchParams.delete('onboardingComplete')
      router.replace(url.pathname + url.search, { scroll: false })

      // Cleanup timeout on unmount
      return () => clearTimeout(confettiTimeout)
    }
  }, [searchParams, router, fireCelebrationConfetti])

  const handleDismiss = () => {
    setShowCelebration(false)
  }

  return (
    <>
      {showCelebration && (
        <OnboardingCompletionBanner onDismiss={handleDismiss} />
      )}
      {children}
    </>
  )
}
