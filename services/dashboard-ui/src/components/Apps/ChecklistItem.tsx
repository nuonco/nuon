'use client'

import React, { type FC } from 'react'
import { Check, CaretDown, CaretRight } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Text } from '@/components/Typography'
import type { TUserJourneyStep } from '@/types'

interface ChecklistItemProps {
  step: TUserJourneyStep
  isExpanded: boolean
  onToggleExpand: () => void
  children?: React.ReactNode
  progressText?: string
}

export const ChecklistItem: FC<ChecklistItemProps> = ({
  step,
  isExpanded,
  onToggleExpand,
  children,
  progressText,
}) => {
  return (
    <div className="border border-gray-200 dark:border-gray-700 rounded-lg">
      {/* Header */}
      <div
        className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50"
        onClick={onToggleExpand}
      >
        <div className="flex items-center gap-3">
          {/* Completion indicator */}
          <div className={`
            flex items-center justify-center w-5 h-5 rounded-full border-2
            ${step.complete
              ? 'bg-green-500 border-green-500 text-white'
              : 'border-gray-300 dark:border-gray-600'
            }
          `}>
            {step.complete && <Check size={12} weight="bold" />}
          </div>

          {/* Title and progress */}
          <div>
            <Text variant="semi-14" className={step.complete ? 'text-gray-600 dark:text-gray-400' : ''}>
              {step.title}
            </Text>
            {progressText && (
              <Text variant="reg-12" className="text-gray-500 dark:text-gray-500 mt-1">
                {progressText}
              </Text>
            )}
          </div>
        </div>

        {/* Expand/collapse button */}
        <Button
          variant="ghost"
          className="!p-1"
          onClick={(e) => {
            e.stopPropagation()
            onToggleExpand()
          }}
        >
          {isExpanded ? <CaretDown size={16} /> : <CaretRight size={16} />}
        </Button>
      </div>

      {/* Expandable content */}
      {isExpanded && children && (
        <div className="border-t border-gray-200 dark:border-gray-700 p-4 bg-gray-50 dark:bg-gray-800/25">
          {children}
        </div>
      )}
    </div>
  )
}