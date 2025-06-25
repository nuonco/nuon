'use client'

import React, { forwardRef, useEffect, useRef } from 'react'
import { cn } from '@/stratus/components/helpers'
import { Button, Icon, Text } from '@/stratus/components/common'
import { useDashboard } from '@/stratus/context'
import './Toast.css'

type TToastVariant = 'default' | 'info' | 'error' | 'warn' | 'success'

export interface IToast
  extends Omit<
    React.HTMLAttributes<HTMLDivElement>,
    'onMouseEnter' | 'onMouseLeave'
  > {
  heading: React.ReactNode
  ref?: React.Ref<HTMLDivElement>
  timeout?: number
  toastId?: string
  variant?: TToastVariant
}

export const Toast = forwardRef<HTMLDivElement, IToast>(
  (
    {
      children,
      className,
      heading,
      timeout = 3000,
      toastId,
      variant = 'default',
      ...props
    },
    ref
  ) => {
    const { removeToast } = useDashboard()
    const timerId = useRef<number | null>(null)

    const handleRemove = () => {
      removeToast(toastId)
    }

    const startTimer = () => {
      clearTimer()
      timerId.current = window.setTimeout(handleRemove, timeout)
    }

    const clearTimer = () => {
      if (timerId.current) {
        clearTimeout(timerId.current)
        timerId.current = null
      }
    }

    useEffect(() => {
      startTimer()
      return clearTimer
    }, [])

    return (
      <div
        className={cn('toast group', variant, className)}
        onMouseEnter={clearTimer}
        onMouseLeave={startTimer}
        ref={ref}
        {...props}
      >
        <div className="flex items-center justify-between">
          <div className="flex gap-4 items-center">
            <Icon variant="RocketLaunch" />
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
          >
            <Icon variant="X" />
          </Button>
        </div>
        <Text className="ml-8 flex flex-col gap-4" variant="subtext">
          {children}
        </Text>
      </div>
    )
  }
)

Toast.displayName = 'Toast'

import { DateTime } from 'luxon'

// example toast
export const ExampleToast = () => {
  const { addToast } = useDashboard()

  return (
    <Button
      onClick={() => {
        const now = DateTime.now()
        addToast(<Toast heading="Cool toast">{now.toFormat('HH:mm:ss')}</Toast>)
      }}
    >
      Default toast
    </Button>
  )
}

export const PermToast = () => {
  const { addToast } = useDashboard()

  return (
    <Button
      onClick={() => {
        const now = DateTime.now()
        addToast(
          <Toast timeout={40000} heading="Cool toast">
            {now.toFormat('HH:mm:ss')}
          </Toast>
        )
      }}
    >
      Long timeout toast
    </Button>
  )
}
