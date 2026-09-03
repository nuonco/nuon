import { useCallback, useEffect, useRef, useState } from 'react'
import { Button } from '@/components/common/Button'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import type { TInstallWorkflowStep } from '@/types'
import { getWorkflowStepTitle } from '@/utils/workflow-utils'
import {
  stepStatusCategory,
  type TStepStatusCategory,
} from '../shared/step-status'

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

const NavButton = ({
  direction,
  onClick,
}: {
  direction: 'left' | 'right'
  onClick: () => void
}) => (
  <Button
    variant="icon"
    aria-label={direction === 'left' ? 'Previous steps' : 'Next steps'}
    onClick={onClick}
    className={cn(
      'absolute top-1/2 z-20 -translate-y-1/2 bg-white dark:bg-dark-grey-700 shadow-sm',
      direction === 'left' ? 'left-1' : 'right-1'
    )}
  >
    <Icon
      variant={direction === 'left' ? 'CaretLeftIcon' : 'CaretRightIcon'}
      size={16}
    />
  </Button>
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
          <div className="pointer-events-none absolute inset-y-0 left-0 z-10 w-12 bg-gradient-to-r from-[var(--background)] to-transparent" />
          <NavButton direction="left" onClick={() => scrollByPage(-1)} />
        </>
      )}

      <div
        ref={viewportRef}
        className="overflow-x-auto overflow-y-hidden snap-x [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden"
      >
        <div className="flex min-w-full overflow-hidden rounded-md border divide-x">
          {steps.map((step, idx) => {
            const category = stepStatusCategory(step.status?.status)
            const isSelected = selectedStepId === step.id

            return (
              <Button
                key={step.id || idx}
                variant="ghost"
                ref={isSelected ? selectedCardRef : undefined}
                aria-current={isSelected ? 'step' : undefined}
                className={cn(
                  'snap-start scroll-mx-12 !h-auto min-h-24 min-w-40 flex-1 basis-40 !items-start !justify-start !rounded-none !px-4 !py-3',
                  isSelected && 'bg-cool-grey-100 dark:bg-dark-grey-700'
                )}
                onClick={() => onSelectStep(step)}
              >
                <Status
                  status={step.status?.status}
                  variant="timeline"
                  isWithoutText
                  iconSize={14}
                />
                <span className="flex min-w-0 flex-1 flex-col items-start gap-1">
                  <Text
                    variant="subtext"
                    weight="strong"
                    className="line-clamp-2 text-left"
                  >
                    {getWorkflowStepTitle(step) || 'Unknown'}
                  </Text>
                  <span className="sr-only">
                    {STEP_STATUS_LABELS[category]}
                  </span>
                  {step.execution_time ? (
                    <Duration
                      nanoseconds={step.execution_time}
                      variant="label"
                      theme="neutral"
                      family="mono"
                    />
                  ) : null}
                </span>
              </Button>
            )
          })}
        </div>
      </div>

      {canScrollRight && (
        <>
          <div className="pointer-events-none absolute inset-y-0 right-0 z-10 w-12 bg-gradient-to-l from-[var(--background)] to-transparent" />
          <NavButton direction="right" onClick={() => scrollByPage(1)} />
        </>
      )}
    </div>
  )
}
