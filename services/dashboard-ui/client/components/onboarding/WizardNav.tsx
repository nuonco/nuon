import React from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
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

  const canGoBack = currentStepIndex > 0
  const currentStepId = steps[currentStepIndex]?.id
  const canGoForward =
    currentStepIndex < steps.length - 1 && !!currentStepId && completedSteps.has(currentStepId)

  return (
    <div className="border-b flex items-center justify-between p-4 md:p-6 bg-white dark:bg-dark-grey-900 z-10">
      <div className="flex-1">
        {canClose && onClose && (
          <Button variant="secondary" onClick={onClose}>
            Close
          </Button>
        )}
      </div>

      <div className="flex items-center gap-3">
        <Button
          variant="ghost"
          className="!p-2"
          onClick={goPrev}
          disabled={!canGoBack}
          aria-label="Previous step"
        >
          <Icon variant="ArrowLeft" size={20} />
        </Button>

        <div className="flex items-center space-x-2">
          {steps.map((step, index) => {
            const isActive = index === currentStepIndex
            const isComplete = completedSteps.has(step.id)
            const canClick = isComplete || index <= currentStepIndex

            return (
              <React.Fragment key={step.id}>
                <Tooltip
                  tipContentClassName="w-max"
                  tipContent={
                    <Text className="!block" variant="subtext">
                      {step.title}
                    </Text>
                  }
                  position="bottom"
                >
                  <button
                    type="button"
                    onClick={() => canClick && goToStep(index)}
                    disabled={!canClick}
                    aria-label={`Go to step ${index + 1}: ${step.title}`}
                    className={cn(
                      'w-5 h-5 rounded-full flex items-center justify-center transition-all duration-300 flex-shrink-0 text-[9px] font-strong leading-none',
                      isActive && 'bg-primary-600 text-white scale-110',
                      isComplete && !isActive && 'bg-green-500 text-white',
                      !isActive &&
                        !isComplete &&
                        'bg-cool-grey-200 dark:bg-cool-grey-700 text-cool-grey-500 dark:text-cool-grey-400',
                      canClick
                        ? 'cursor-pointer hover:scale-125'
                        : 'cursor-not-allowed'
                    )}
                  >
                    {isComplete && !isActive ? (
                      <Icon variant="Check" size={10} weight="bold" />
                    ) : (
                      index + 1
                    )}
                  </button>
                </Tooltip>

                {index < steps.length - 1 && (
                  <div
                    className={cn(
                      'h-0.5 w-6 rounded-full transition-all duration-300 flex-shrink-0',
                      isComplete
                        ? 'bg-green-500'
                        : 'bg-cool-grey-200 dark:bg-cool-grey-700'
                    )}
                  />
                )}
              </React.Fragment>
            )
          })}
        </div>

        <Button
          variant="ghost"
          className="!p-2"
          onClick={goNext}
          disabled={!canGoForward}
          aria-label="Next step"
        >
          <Icon variant="ArrowRight" size={20} />
        </Button>
      </div>

      <div className="flex-1" />
    </div>
  )
}
