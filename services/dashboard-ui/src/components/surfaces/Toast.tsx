'use client'

import React, { forwardRef, useEffect, useRef } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useToast } from '@/hooks/use-toast'
import { cn } from '@/utils/classnames'
import './Toast.css'

export type TToastTheme = 'default' | 'info' | 'error' | 'warn' | 'success'

export interface IToast
  extends Omit<
    React.HTMLAttributes<HTMLDivElement>,
    'onMouseEnter' | 'onMouseLeave'
  > {
  heading: React.ReactNode
  pauseTimeout?: boolean
  ref?: React.Ref<HTMLDivElement>
  timeout?: number
  toastId?: string
  theme?: TToastTheme
}

export const THEME_CLASSES = {
  default:
    'bg-cool-grey-50 text-dark-grey-950 dark:bg-dark-grey-800 dark:text-white',
  success:
    'bg-green-100/50 text-green-800 !border-green-400 dark:bg-[#0C1B14] dark:!border-green-500/40 dark:text-green-500',
  warn: 'bg-orange-100/50 text-orange-800 !border-orange-400 dark:bg-[#2E1E10] dark:!border-orange-500/40 dark:text-orange-500',
  error:
    'bg-red-100/50 text-red-800 !border-red-300 dark:bg-[#290C0D] dark:!border-red-500/40 dark:text-red-500',
  info: 'bg-blue-100/50 text-blue-800 !border-blue-400 dark:bg-[#0F172A] dark:!border-blue-500/40 dark:text-blue-500',
}

// Helper function to determine the appropriate role and aria-live based on theme
const getToastAccessibilityProps = (theme: TToastTheme) => {
  switch (theme) {
    case 'error':
      return {
        role: 'alert' as const,
        'aria-live': 'assertive' as const,
      }
    case 'warn':
      return {
        role: 'alert' as const,
        'aria-live': 'assertive' as const,
      }
    case 'success':
    case 'info':
    case 'default':
      return {
        role: 'status' as const,
        'aria-live': 'polite' as const,
      }
  }
}

export const Toast = forwardRef<HTMLDivElement, IToast>(
  (
    {
      children,
      className,
      heading,
      pauseTimeout = false,
      timeout = 5000,
      toastId,
      theme = 'default',
      ...props
    },
    ref
  ) => {
    const { removeToast } = useToast()
    const timerId = useRef<number | null>(null)
    const start = useRef<number>(Date.now())
    const remaining = useRef<number>(timeout)

    const handleRemove = () => {
      removeToast(toastId as string)
    }

    const startTimer = () => {
      clearTimer()
      timerId.current = window.setTimeout(handleRemove, remaining?.current)
    }

    const clearTimer = () => {
      if (timerId.current) {
        clearTimeout(timerId.current)
        timerId.current = null
      }
    }

    useEffect(() => {
      if (pauseTimeout) {
        remaining.current = Date.now() - start?.current
        clearTimer()
      } else {
        startTimer()
      }

      return clearTimer
    }, [pauseTimeout])

    const accessibilityProps = getToastAccessibilityProps(theme)

    return (
      <div
        className={cn(
          'toast absolute bottom-0 right-0 z-10 group flex flex-col border rounded-md w-82 p-4 gap-1 shadow-md',
          THEME_CLASSES[theme],
          className
        )}
        ref={ref}
        {...accessibilityProps}
        {...props}
      >
        <div className="flex items-center justify-between">
          <div className="flex gap-4 items-center">
            {typeof heading === 'string' ? (
              <Text weight="strong">{heading}</Text>
            ) : (
              heading
            )}
          </div>
          <Button
            className="!p-1 !h-auto opacity-0 group-hover:opacity-100 transition-opacity"
            onClick={handleRemove}
            variant="ghost"
            aria-label="Close notification"
            title="Close notification"
          >
            <Icon variant="X" aria-hidden="true" />
          </Button>
        </div>
        <Text className="flex flex-col gap-4" variant="subtext">
          {children}
        </Text>
      </div>
    )
  }
)

Toast.displayName = 'Toast'
