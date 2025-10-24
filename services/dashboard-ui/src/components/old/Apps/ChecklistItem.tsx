'use client'

import React, { type FC } from 'react'
import { Check, CaretDown, CaretRight } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { Text } from '@/components/old/Typography'
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
    <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden transition-all duration-200 ease-out">
      {/* Header */}
      <div
        className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors duration-150"
        onClick={onToggleExpand}
      >
        <div className="flex items-center gap-3">
          {/* Completion indicator with scale transition */}
          <div
            className={`
            flex items-center justify-center w-5 h-5 rounded-full border-2 transition-all duration-200 ease-out transform
            ${
              step.complete
                ? 'bg-green-500 border-green-500 text-white scale-110'
                : 'border-gray-300 dark:border-gray-600 scale-100'
            }
          `}
          >
            {step.complete && (
              <Check
                size={12}
                weight="bold"
                className="transition-all duration-150 ease-out"
              />
            )}
          </div>

          {/* Title and progress */}
          <div>
            <Text
              variant="semi-14"
              className={`transition-colors duration-200 ${step.complete ? 'text-gray-600 dark:text-gray-400' : ''}`}
            >
              {step.title}
            </Text>
            {progressText && (
              <Text
                variant="reg-12"
                className="text-gray-500 dark:text-gray-500 mt-1"
              >
                {progressText}
              </Text>
            )}
          </div>
        </div>

        {/* Expand/collapse button with rotation transition */}
        <Button
          variant="ghost"
          className="!p-1 transition-transform duration-200 ease-out hover:scale-110"
          onClick={(e) => {
            e.stopPropagation()
            onToggleExpand()
          }}
        >
          <CaretDown
            size={16}
            className={`transition-transform duration-200 ease-out ${
              isExpanded ? 'rotate-0' : '-rotate-90'
            }`}
          />
        </Button>
      </div>

      {/* Expandable content with height transition */}
      <div
        className={`transition-all duration-300 ease-out overflow-hidden ${
          isExpanded && children ? 'opacity-100' : 'max-h-0 opacity-0'
        }`}
      >
        {children && (
          <div className="border-t border-gray-200 dark:border-gray-700 p-4 bg-gray-50 dark:bg-gray-800/25">
            <div
              className={`transition-all duration-200 ease-out ${
                isExpanded ? 'translate-y-0' : '-translate-y-2'
              }`}
            >
              {children}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
