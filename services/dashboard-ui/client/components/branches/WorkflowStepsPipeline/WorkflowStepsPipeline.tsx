import { useCallback, useEffect, useRef, useState } from 'react'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import type { TInstallWorkflowStep } from '@/types'
import { getWorkflowStepTitle } from '@/utils/workflow-utils'
import { stepStatusCategory, type TStepStatusCategory } from '../shared/step-status'

interface IWorkflowStepsPipeline {
  steps: TInstallWorkflowStep[]
  selectedStepId?: string
  onSelectStep: (step: TInstallWorkflowStep) => void
}

const STEP_STATUS_LABELS: Record<TStepStatusCategory, string> = {
  success: 'Completed',
  error: 'Failed',
  active: 'In progress',
  awaiting: 'Awaiting approval',
  pending: 'Pending',
}

const StatusIcon = ({ category }: { category: TStepStatusCategory }) => {
  if (category === 'success') {
    return (
      <div
        aria-hidden
        className="w-[26px] h-[26px] rounded-full bg-green-500 flex items-center justify-center shrink-0"
      >
        <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
          <path d="M2.5 6.5L5.5 9.5L10.5 4" stroke="white" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>
    )
  }

  if (category === 'error') {
    return (
      <div
        aria-hidden
        className="w-[26px] h-[26px] rounded-full bg-red-500 flex items-center justify-center shrink-0"
      >
        <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
          <path d="M4 4L9 9M9 4L4 9" stroke="white" strokeWidth="1.8" strokeLinecap="round" />
        </svg>
      </div>
    )
  }

  if (category === 'active') {
    return (
      <div
        aria-hidden
        className="w-[26px] h-[26px] rounded-full bg-blue-500 flex items-center justify-center shrink-0"
      >
        <svg className="animate-spin" width="16" height="16" viewBox="0 0 16 16" fill="none">
          <circle cx="8" cy="8" r="6" stroke="white" strokeOpacity="0.3" strokeWidth="2" />
          <path d="M8 2 A6 6 0 0 1 14 8" stroke="white" strokeWidth="2" strokeLinecap="round" />
        </svg>
      </div>
    )
  }

  if (category === 'awaiting') {
    return (
      <div
        aria-hidden
        className="w-[26px] h-[26px] rounded-full bg-amber-500 flex items-center justify-center shrink-0"
      >
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <circle cx="7" cy="7" r="5.25" stroke="white" strokeWidth="1.5" />
          <path d="M7 4.25V7L8.75 8.25" stroke="white" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>
    )
  }

  return (
    <div
      aria-hidden
      className="w-[26px] h-[26px] rounded-full flex items-center justify-center shrink-0 ring-1 ring-inset ring-cool-grey-400/40 dark:ring-dark-grey-500/40"
    >
      <div className="w-[5px] h-[5px] rounded-full bg-cool-grey-400 dark:bg-dark-grey-500" />
    </div>
  )
}

const Arrow = ({ filled }: { filled: boolean }) => (
  <svg
    aria-hidden
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

const NavButton = ({
  direction,
  onClick,
}: {
  direction: 'left' | 'right'
  onClick: () => void
}) => (
  <button
    type="button"
    aria-label={direction === 'left' ? 'Previous steps' : 'Next steps'}
    onClick={onClick}
    className={cn(
      'absolute top-1/2 z-20 -translate-y-1/2 flex h-8 w-8 items-center justify-center rounded-full',
      'border bg-white dark:bg-dark-grey-800 shadow-sm',
      'text-cool-grey-600 dark:text-cool-grey-300 hover:brightness-105',
      direction === 'left' ? 'left-1' : 'right-1'
    )}
  >
    <Icon variant={direction === 'left' ? 'CaretLeftIcon' : 'CaretRightIcon'} size={16} />
  </button>
)

export const WorkflowStepsPipeline = ({
  steps,
  selectedStepId,
  onSelectStep,
}: IWorkflowStepsPipeline) => {
  const viewportRef = useRef<HTMLDivElement>(null)
  const selectedCardRef = useRef<HTMLButtonElement>(null)
  const [canScrollLeft, setCanScrollLeft] = useState(false)
  const [canScrollRight, setCanScrollRight] = useState(false)

  const updateScrollState = useCallback(() => {
    const el = viewportRef.current
    if (!el) return
    setCanScrollLeft(el.scrollLeft > 1)
    setCanScrollRight(el.scrollLeft < el.scrollWidth - el.clientWidth - 1)
  }, [])

  useEffect(() => {
    const el = viewportRef.current
    if (!el) return
    updateScrollState()
    el.addEventListener('scroll', updateScrollState, { passive: true })
    const observer = new ResizeObserver(updateScrollState)
    observer.observe(el)
    return () => {
      el.removeEventListener('scroll', updateScrollState)
      observer.disconnect()
    }
  }, [updateScrollState, steps.length])

  useEffect(() => {
    const viewport = viewportRef.current
    const card = selectedCardRef.current
    if (!viewport || !card) return

    // Only ever move this strip's scrollLeft: native scrollIntoView also scrolls
    // overflow-hidden ancestors (the surrounding panel), shifting the whole page.
    const viewportBox = viewport.getBoundingClientRect()
    const cardBox = card.getBoundingClientRect()
    const delta =
      cardBox.left +
      cardBox.width / 2 -
      (viewportBox.left + viewportBox.width / 2)

    if (Math.abs(delta) < 1) return
    viewport.scrollTo({ left: viewport.scrollLeft + delta, behavior: 'smooth' })
  }, [selectedStepId])

  const scrollByPage = (dir: 1 | -1) => {
    const el = viewportRef.current
    if (!el) return
    el.scrollBy({ left: dir * el.clientWidth * 0.85, behavior: 'smooth' })
  }

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
    <div className="relative">
      {canScrollLeft && (
        <>
          <div className="pointer-events-none absolute inset-y-0 left-0 z-10 w-12 bg-gradient-to-r from-white dark:from-dark-grey-900 to-transparent" />
          <NavButton direction="left" onClick={() => scrollByPage(-1)} />
        </>
      )}

      <div
        ref={viewportRef}
        className="overflow-x-auto overflow-y-hidden snap-x [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden"
      >
        <div className="flex items-stretch gap-2 py-2 px-1 min-w-max">
          {steps.map((step, idx) => {
            const category = stepStatusCategory(step.status?.status)
            const isSelected = selectedStepId === step.id
            const prevStep = idx > 0 ? steps[idx - 1] : null
            const prevSuccess = stepStatusCategory(prevStep?.status?.status) === 'success'

            let cardBorder = ''
            let cardBg = 'bg-cool-grey-50 dark:bg-dark-grey-800'
            let cardRing = ''

            if (isSelected) {
              cardBorder = 'border-primary-500 dark:border-primary-400'
              cardBg = 'bg-primary-50 dark:bg-primary-500/15'
              cardRing = 'ring-2 ring-primary-500/20'
            } else if (category === 'active') {
              cardBorder = 'border-blue-400/50 dark:border-blue-500/50'
              cardBg = 'bg-blue-50/40 dark:bg-blue-500/10'
            } else if (category === 'awaiting') {
              cardBorder = 'border-amber-400/50 dark:border-amber-500/50'
              cardBg = 'bg-amber-50/40 dark:bg-amber-500/10'
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

                <button
                  type="button"
                  ref={isSelected ? selectedCardRef : undefined}
                  aria-current={isSelected ? 'step' : undefined}
                  className={cn(
                    'snap-start scroll-mx-12 flex flex-col flex-1 min-w-[168px] items-center justify-center gap-2 px-4 py-4 rounded-[10px] cursor-pointer border transition-all hover:brightness-105',
                    'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-400/80',
                    cardBorder,
                    cardBg,
                    cardRing
                  )}
                  onClick={() => onSelectStep(step)}
                >
                  <StatusIcon category={category} />

                  <Text
                    variant="body"
                    weight="strong"
                    className="text-center leading-tight max-w-[160px] text-cool-grey-900 dark:text-cool-grey-100"
                  >
                    {getWorkflowStepTitle(step) || 'Unknown'}
                  </Text>

                  <span className="sr-only">{STEP_STATUS_LABELS[category]}</span>

                  {step.execution_time ? (
                    <Duration
                      nanoseconds={step.execution_time}
                      variant="label"
                      theme="neutral"
                      family="mono"
                    />
                  ) : null}
                </button>
              </div>
            )
          })}
        </div>
      </div>

      {canScrollRight && (
        <>
          <div className="pointer-events-none absolute inset-y-0 right-0 z-10 w-12 bg-gradient-to-l from-white dark:from-dark-grey-900 to-transparent" />
          <NavButton direction="right" onClick={() => scrollByPage(1)} />
        </>
      )}
    </div>
  )
}
