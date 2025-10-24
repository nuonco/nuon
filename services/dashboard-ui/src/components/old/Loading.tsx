'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { Text } from '@/components/old/Typography'

export const SpinnerSVG: FC<{ variant?: 'default' | 'large' }> = ({
  variant = 'default',
}) => {
  return (
    <span className="animate-pulse">
      <svg
        className={classNames('animate-spin', {
          'h-5 w-5': variant === 'default',
          'h-10 w-10': variant === 'large',
        })}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          className="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          strokeWidth="4"
        ></circle>
        <path
          className="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        ></path>
      </svg>
    </span>
  )
}

export interface ILoading {
  loadingText?: string
  textClassName?: string
  variant?: 'default' | 'page' | 'stack'
}

export const Loading: FC<ILoading> = ({
  loadingText = 'Getting things ready...',
  textClassName = '',
  variant = 'default',
}) => {
  return (
    <div
      className={classNames('flex', {
        'h-[calc(100vh-350px)] m-auto': variant === 'page',
      })}
    >
      <span
        className={classNames('flex items-center', {
          'flex-col gap-4 m-auto': variant === 'page' || variant === 'stack',
          'gap-2': variant === 'default',
        })}
      >
        <SpinnerSVG
          variant={
            variant === 'page' || variant === 'stack' ? 'large' : 'default'
          }
        />
        <Text className={textClassName} variant="reg-14">
          {loadingText}
        </Text>
      </span>
    </div>
  )
}
