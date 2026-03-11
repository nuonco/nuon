import React from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'
import { cn } from '@/utils/classnames'

export function WizardNav() {
  const {
    steps,
    currentStepIndex,
    completedSteps,
    canClose,
    onClose,
    goNext,
    goPrev,
    goToStep,
  } = useOnboardingWizard()

  const { isStepComplete, getStepMetadata } = useOnboardingJourney()

  const orgCreated = isStepComplete('org_created')
  const orgId = getStepMetadata('org_created', 'org_id') as string | undefined

  const canGoBack = currentStepIndex > 0
  const currentStepId = steps[currentStepIndex]?.id
  const canGoForward =
    currentStepIndex < steps.length - 1 && !!currentStepId && completedSteps.has(currentStepId)

  return (
    <div className="flex items-center gap-4 px-6 pt-8 pb-6 bg-white dark:bg-dark-grey-900 z-10">
      <div className="w-16 flex-shrink-0">
        {canClose && onClose && (
          <Button variant="secondary" onClick={onClose}>
            Close
          </Button>
        )}
      </div>

      <div className="flex items-center gap-2 flex-1 min-w-0">
        <Button
          variant="ghost"
          className="!p-2 flex-shrink-0"
          onClick={goPrev}
          disabled={!canGoBack}
          aria-label="Previous step"
        >
          <Icon variant="ArrowLeft" size={20} />
        </Button>

        <div className="flex items-center w-full max-w-2xl mx-auto">
          {steps.map((step, index) => {
            const isActive = index === currentStepIndex
            const isComplete = completedSteps.has(step.id)
            const canClick = isComplete || index <= currentStepIndex
            const isLast = index === steps.length - 1

            return (
              <React.Fragment key={step.id}>
                <div className="relative flex-shrink-0">
                  <button
                    type="button"
                    onClick={() => canClick && goToStep(index)}
                    disabled={!canClick}
                    aria-label={`Go to step ${index + 1}: ${step.title}`}
                    className={cn(
                      'w-[26px] h-[26px] rounded-full flex items-center justify-center transition-colors duration-300',
                      isComplete && !isActive && 'bg-primary-600',
                      isActive &&
                        'bg-primary-100 dark:bg-primary-950 border-2 border-primary-600 dark:border-primary-400',
                      !isActive &&
                        !isComplete &&
                        'bg-cool-grey-500/[0.16] border border-cool-grey-500/25',
                      canClick ? 'cursor-pointer' : 'cursor-not-allowed'
                    )}
                  >
                    {isComplete && !isActive ? (
                      <Icon
                        variant="Check"
                        size={14}
                        weight="bold"
                        className="text-white"
                      />
                    ) : isActive ? (
                      <div className="w-2.5 h-2.5 rounded-full bg-primary-600 dark:bg-primary-400" />
                    ) : (
                      <div className="w-1.5 h-1.5 rounded-full bg-cool-grey-500 dark:bg-cool-grey-600" />
                    )}
                  </button>
                  <Text
                    variant="label"
                    weight="strong"
                    className={cn(
                      'absolute top-full mt-1.5 left-1/2 -translate-x-1/2 whitespace-nowrap text-center',
                      isActive && 'text-primary-600 dark:text-primary-400',
                      isComplete &&
                        !isActive &&
                        'text-cool-grey-600 dark:text-white/70',
                      !isActive &&
                        !isComplete &&
                        'text-cool-grey-500 dark:text-cool-grey-400'
                    )}
                  >
                    {step.navLabel ?? step.title}
                  </Text>
                </div>

                {!isLast && (
                  <div
                    className={cn(
                      'h-0.5 flex-1 min-w-3 rounded-full transition-colors duration-300',
                      isComplete
                        ? 'bg-primary-600'
                        : 'bg-cool-grey-500 dark:bg-cool-grey-700'
                    )}
                  />
                )}
              </React.Fragment>
            )
          })}
        </div>

        <div className="flex items-center gap-1 flex-shrink-0">
          <Button
            variant="ghost"
            className="!p-2"
            onClick={goNext}
            disabled={!canGoForward}
            aria-label="Next step"
          >
            <Icon variant="ArrowRight" size={20} />
          </Button>
          <Text variant="label" theme="neutral" className="w-10 tabular-nums">
            {currentStepIndex + 1}/{steps.length}
          </Text>
        </div>
      </div>

      <div className="w-16 flex-shrink-0 flex justify-end">
        {orgCreated && orgId && (
          <Button variant="ghost" href={`/${orgId}/apps`}>
            Exit
          </Button>
        )}
      </div>
    </div>
  )
}
