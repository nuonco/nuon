import { useEffect, useRef } from 'react'
import { Badge } from '@/components/common/Badge'
import { Duration } from '@/components/common/Duration'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import type { TInstallWorkflowStep } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { stepStatusCategory, type TStepStatusCategory } from '../shared/step-status'

interface IWorkflowStepsPipeline {
  steps: TInstallWorkflowStep[]
  selectedStepId?: string
  onSelectStep: (step: TInstallWorkflowStep) => void
}

const StatusIcon = ({ category }: { category: TStepStatusCategory }) => {
  if (category === 'success') {
    return (
      <div className="w-[26px] h-[26px] rounded-full bg-green-500 flex items-center justify-center shrink-0">
        <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
          <path d="M2.5 6.5L5.5 9.5L10.5 4" stroke="white" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>
    )
  }

  if (category === 'error') {
    return (
      <div className="w-[26px] h-[26px] rounded-full bg-red-500 flex items-center justify-center shrink-0">
        <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
          <path d="M4 4L9 9M9 4L4 9" stroke="white" strokeWidth="1.8" strokeLinecap="round" />
        </svg>
      </div>
    )
  }

  if (category === 'active') {
    return (
      <div className="w-[26px] h-[26px] rounded-full bg-blue-500 flex items-center justify-center shrink-0">
        <svg className="animate-spin" width="16" height="16" viewBox="0 0 16 16" fill="none">
          <circle cx="8" cy="8" r="6" stroke="rgba(255,255,255,0.3)" strokeWidth="2" />
          <path d="M8 2 A6 6 0 0 1 14 8" stroke="white" strokeWidth="2" strokeLinecap="round" />
        </svg>
      </div>
    )
  }

  if (category === 'awaiting') {
    return (
      <div className="w-[26px] h-[26px] rounded-full bg-amber-500 flex items-center justify-center shrink-0">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <circle cx="7" cy="7" r="5.25" stroke="white" strokeWidth="1.5" />
          <path d="M7 4.25V7L8.75 8.25" stroke="white" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>
    )
  }

  return (
    <div
      className="w-[26px] h-[26px] rounded-full flex items-center justify-center shrink-0"
      style={{ boxShadow: 'inset 0 0 0 1.5px rgba(150,150,170,0.35)' }}
    >
      <div className="w-[5px] h-[5px] rounded-full bg-cool-grey-400 dark:bg-dark-grey-500" />
    </div>
  )
}

const Arrow = ({ filled }: { filled: boolean }) => (
  <svg
    width="20"
    height="20"
    viewBox="0 0 20 20"
    fill="none"
    className={`shrink-0 self-center transition-colors ${filled ? 'text-green-500' : 'text-cool-grey-300 dark:text-cool-grey-600'}`}
  >
    <path
      d="M4 10H16M16 10L11 5M16 10L11 15"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

export const WorkflowStepsPipeline = ({
  steps,
  selectedStepId,
  onSelectStep,
}: IWorkflowStepsPipeline) => {
  const selectedCardRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    selectedCardRef.current?.scrollIntoView({ behavior: 'smooth', inline: 'nearest', block: 'nearest' })
  }, [selectedStepId])

  if (steps.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 gap-4">
        <Loading variant="large" />
        <Text variant="body" theme="neutral">
          Generating workflow steps...
        </Text>
      </div>
    )
  }

  return (
    <div
      className="relative overflow-x-auto overflow-y-hidden"
      style={{ scrollbarWidth: 'thin', scrollBehavior: 'smooth' }}
    >
      <div className="flex items-stretch gap-2 py-2 px-1 min-w-max">
        {steps.map((step, idx) => {
          const category = stepStatusCategory(step.status?.status)
          const isSelected = selectedStepId === step.id
          const prevStep = idx > 0 ? steps[idx - 1] : null
          const prevSuccess = stepStatusCategory(prevStep?.status?.status) === 'success'

          let cardBorder = ''
          let cardBg = 'bg-cool-grey-50 dark:bg-dark-grey-800'
          let cardShadow = ''

          if (isSelected) {
            cardBorder = 'border-primary-500 dark:border-primary-400'
            cardBg = 'bg-primary-50 dark:bg-[#1B1026]'
            cardShadow = '0 0 0 3px rgba(124,58,237,0.18)'
          } else if (category === 'active') {
            cardBorder = 'border-blue-400/50 dark:border-blue-500/50'
            cardBg = 'bg-blue-50/40 dark:bg-[#0d1b2e]'
          } else if (category === 'awaiting') {
            cardBorder = 'border-amber-400/50 dark:border-amber-500/50'
            cardBg = 'bg-amber-50/40 dark:bg-[#2a1f06]'
          } else if (category === 'success') {
            cardBorder = 'border-green-400/50 dark:border-green-500/40'
            cardBg = 'bg-green-50/30 dark:bg-dark-grey-800'
          } else if (category === 'error') {
            cardBorder = 'border-red-400/50 dark:border-red-500/40'
            cardBg = 'bg-red-50/30 dark:bg-dark-grey-800'
          }

          return (
            <div key={step.id || idx} className="flex items-stretch gap-2 flex-1 min-w-0">
              {idx > 0 && <Arrow filled={prevSuccess} />}

              <div
                ref={isSelected ? selectedCardRef : undefined}
                className={`flex flex-col flex-1 min-w-[168px] items-center justify-center gap-2 px-4 py-4 rounded-[10px] cursor-pointer border transition-all ${cardBorder} ${cardBg} hover:brightness-105`}
                style={cardShadow ? { boxShadow: cardShadow } : undefined}
                onClick={() => onSelectStep(step)}
              >
                <StatusIcon category={category} />

                <Text
                  variant="body"
                  weight="strong"
                  className="text-center leading-tight max-w-[160px] text-cool-grey-900 dark:text-cool-grey-100"
                >
                  {toSentenceCase(step.name || 'Unknown')}
                </Text>

                <div className="flex items-center gap-2">
                  <Badge size="sm" variant="code">
                    GROUP {step.group_idx ?? idx + 1}
                  </Badge>
                  {step.execution_time ? (
                    <Duration
                      nanoseconds={step.execution_time}
                      variant="label"
                      theme="neutral"
                      family="mono"
                    />
                  ) : null}
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
