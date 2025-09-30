'use client'

import { useEffect, useRef, useCallback } from 'react'
import JSConfetti from 'js-confetti'

export const useConfetti = () => {
  const confettiRef = useRef<JSConfetti | null>(null)

  useEffect(() => {
    confettiRef.current = new JSConfetti()

    return () => {
      confettiRef.current = null
    }
  }, [])

  const fireConfetti = useCallback(() => {
    if (confettiRef.current) {
      confettiRef.current.addConfetti({
        confettiColors: ['#10B981', '#3B82F6', '#F59E0B', '#EF4444', '#8B5CF6'],
        confettiRadius: 4,
        confettiNumber: 200,
      })
    }
  }, [])

  const fireCelebrationConfetti = useCallback(() => {
    if (confettiRef.current) {
      confettiRef.current.addConfetti({
        confettiColors: ['#10B981', '#059669', '#34D399', '#F59E0B', '#3B82F6'],
        confettiRadius: 5,
        confettiNumber: 250,
      })
    }
  }, [])

  return { fireConfetti, fireCelebrationConfetti }
}

