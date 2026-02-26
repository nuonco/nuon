'use client'

import { createContext, useEffect, useRef, useState, ReactNode } from 'react'

interface IAutoRefreshContextType {
  timeRemaining: number
  resetTimer: () => void
  pauseTimer: () => void
  resumeTimer: () => void
  isPaused: boolean
}

export const AutoRefreshContext = createContext<
  IAutoRefreshContextType | undefined
>(undefined)

interface IAutoRefreshProviderProps {
  children: ReactNode
  refreshIntervalMs?: number // Default to 5 minutes
  showWarning?: boolean // Show warning before refresh
  warningTimeMs?: number // Show warning X ms before refresh
}

export function AutoRefreshProvider({
  children,
  refreshIntervalMs = 5 * 60 * 1000, // 5 minutes
  showWarning = false,
  warningTimeMs = 30 * 1000, // 30 seconds before refresh
}: IAutoRefreshProviderProps) {
  const timeoutRef = useRef<NodeJS.Timeout | null>(null)
  const warningTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const startTimeRef = useRef<number>(Date.now())
  const pausedTimeRef = useRef<number>(0)
  const isPausedRef = useRef<boolean>(false)

  const [timeRemaining, setTimeRemaining] = useState<number>(refreshIntervalMs)
  const [isPaused, setIsPaused] = useState<boolean>(false)

  const hardRefresh = () => {
    // Force a hard refresh by reloading the page
    window.location.reload()
  }

  const showRefreshWarning = () => {
    if (showWarning && !isPausedRef.current) {
      const confirmed = window.confirm(
        `This page will refresh in ${warningTimeMs / 1000} seconds to ensure you have the latest data. Click OK to refresh now or Cancel to continue.`
      )
      if (confirmed) {
        hardRefresh()
      }
    }
  }

  const startTimer = () => {
    clearTimers()
    startTimeRef.current = Date.now()
    isPausedRef.current = false
    setIsPaused(false)

    // Set warning timer if enabled
    if (showWarning) {
      warningTimeoutRef.current = setTimeout(() => {
        if (!isPausedRef.current) {
          showRefreshWarning()
        }
      }, refreshIntervalMs - warningTimeMs)
    }

    // Set main refresh timer
    timeoutRef.current = setTimeout(() => {
      if (!isPausedRef.current) {
        hardRefresh()
      }
    }, refreshIntervalMs)
  }

  const clearTimers = () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
      timeoutRef.current = null
    }
    if (warningTimeoutRef.current) {
      clearTimeout(warningTimeoutRef.current)
      warningTimeoutRef.current = null
    }
  }

  const resetTimer = () => {
    startTimer()
  }

  const pauseTimer = () => {
    if (!isPausedRef.current) {
      pausedTimeRef.current = Date.now()
      isPausedRef.current = true
      setIsPaused(true)
      clearTimers()
    }
  }

  const resumeTimer = () => {
    if (isPausedRef.current) {
      const pausedDuration = Date.now() - pausedTimeRef.current
      const elapsed = pausedTimeRef.current - startTimeRef.current
      const remaining = refreshIntervalMs - elapsed

      if (remaining > 0) {
        isPausedRef.current = false
        setIsPaused(false)
        startTimeRef.current = Date.now() - elapsed

        // Set warning timer for remaining time
        if (showWarning && remaining > warningTimeMs) {
          warningTimeoutRef.current = setTimeout(() => {
            if (!isPausedRef.current) {
              showRefreshWarning()
            }
          }, remaining - warningTimeMs)
        }

        // Set main timer for remaining time
        timeoutRef.current = setTimeout(() => {
          if (!isPausedRef.current) {
            hardRefresh()
          }
        }, remaining)
      } else {
        // Time already expired, refresh immediately
        hardRefresh()
      }
    }
  }

  // Update time remaining counter
  useEffect(() => {
    if (isPausedRef.current) return

    const interval = setInterval(() => {
      if (!isPausedRef.current) {
        const elapsed = Date.now() - startTimeRef.current
        const remaining = Math.max(0, refreshIntervalMs - elapsed)
        setTimeRemaining(remaining)

        if (remaining === 0) {
          clearInterval(interval)
        }
      }
    }, 1000)

    return () => clearInterval(interval)
  }, [refreshIntervalMs, isPaused])

  useEffect(() => {
    startTimer()

    // Cleanup on unmount
    return () => {
      clearTimers()
    }
  }, [refreshIntervalMs])

  const contextValue: IAutoRefreshContextType = {
    timeRemaining,
    resetTimer,
    pauseTimer,
    resumeTimer,
    isPaused,
  }

  return (
    <AutoRefreshContext.Provider value={contextValue}>
      {children}
    </AutoRefreshContext.Provider>
  )
}
