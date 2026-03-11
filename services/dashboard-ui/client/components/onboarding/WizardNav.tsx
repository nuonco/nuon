import React from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'
import { cn } from '@/utils/classnames'

export function WizardNav({ isScrolled = false }: { isScrolled?: boolean }) {
  const {
    steps,
    currentStepIndex,
    completedSteps,
    goToStep,
  } = useOnboardingWizard()

  return (
    <div className={cn(
      'flex items-center px-6 pt-8 pb-7 bg-white dark:bg-dark-grey-900 z-10 transition-shadow duration-200',
      isScrolled && 'shadow-sm'
    )}>
      <div className="flex items-center flex-1 min-w-0 max-w-2xl mx-auto">
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
                <div className="relative flex-1 min-w-3 h-0.5 rounded-full bg-cool-grey-200 dark:bg-cool-grey-700 overflow-hidden">
                  <div
                    className="absolute inset-y-0 left-0 rounded-full bg-primary-600"
                    style={{
                      width: isComplete ? '100%' : '0%',
                      transition: 'width 800ms cubic-bezier(0.65, 0, 0.35, 1)',
                    }}
                  />
                </div>
              )}
            </React.Fragment>
          )
        })}
      </div>
    </div>
  )
}
