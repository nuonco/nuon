'use client'

import React, { forwardRef, useEffect, useRef } from 'react'
import { cn } from '@/stratus/components/helpers'
import { Button, Icon, Text, Link } from '@/stratus/components/common'
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

export const ReprovisionToast = () => {
  const { addToast } = useDashboard()

  return (
    <Button
      variant="ghost"
      onClick={() => {
        addToast(
          <Toast timeout={50000} heading="Reprovision workflow running">
            <span className="flex w-full justify-between flex-wrap">
              <span className="flex flex-col gap-0 w-full">
                <span className="flex items-center gap-4">
                  <progress
                    className="rounded-lg [&::-webkit-progress-bar]:rounded-lg [&::-webkit-progress-value]:rounded-lg   [&::-webkit-progress-bar]:bg-cool-grey-300 [&::-webkit-progress-value]:bg-green-400 [&::-moz-progress-bar]:bg-green-400 [&::-webkit-progress-value]:transition-all [&::-webkit-progress-value]:duration-500 [&::-moz-progress-bar]:transition-all [&::-moz-progress-bar]:duration-500 h-[8px] w-full"
                    max={26}
                    value={17}
                  />
                </span>

                <Text
                  className="flex justify-between w-full gap-4"
                  variant="subtext"
                  theme="muted"
                >
                  <span>
                    {17} of {26} steps done
                  </span>{' '}
                  <Link className="flex items-center" href="#">
                    View details <Icon variant="CaretRight" size="14" />
                  </Link>
                </Text>
              </span>
            </span>
          </Toast>
        )
      }}
    >
      Reprovision install <Icon variant="ArrowURightUp" />
    </Button>
  )
}
