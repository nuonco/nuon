import { useState, useEffect, useRef } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'

export function WizardStepView() {
  const { steps, currentStepIndex, completedSteps, sharedData, setSharedData, goNext, goPrev, markComplete } =
    useOnboardingWizard()

  const [visibleIndex, setVisibleIndex] = useState(currentStepIndex)
  const [isTransitioning, setIsTransitioning] = useState(false)
  const prevIndexRef = useRef(currentStepIndex)

  useEffect(() => {
    if (currentStepIndex === prevIndexRef.current) return

    setIsTransitioning(true)
    const timer = setTimeout(() => {
      setVisibleIndex(currentStepIndex)
      prevIndexRef.current = currentStepIndex
      setIsTransitioning(false)
    }, 150)

    return () => clearTimeout(timer)
  }, [currentStepIndex])

  const visibleStep = steps[visibleIndex]
  if (!visibleStep) return null

  const StepComponent = visibleStep.component
  const canGoBack = currentStepIndex > 0

  return (
    <div
      className={`transition-all duration-150 ease-out ${
        isTransitioning ? 'opacity-0 translate-x-2' : 'opacity-100 translate-x-0'
      }`}
    >
      <div className="mb-6">
        {canGoBack && (
          <Button variant="secondary" size="sm" onClick={goPrev}>
            <Icon variant="ArrowLeft" size={14} />
            Back
          </Button>
        )}
      </div>
      <div className="mb-8">
        <Text variant="h2" role="heading" level={2} className="mb-2">
          {visibleStep.title}
        </Text>
        {visibleStep.description && (
          <Text variant="body" theme="neutral" className="!block max-w-md !text-pretty">
            {visibleStep.description}
          </Text>
        )}
      </div>

      <StepComponent
        isComplete={completedSteps.has(visibleStep.id)}
        sharedData={sharedData}
        setSharedData={setSharedData}
        onAdvance={() => {
          markComplete(visibleStep.id)
          goNext()
        }}
        nextStepTitle={steps[visibleIndex + 1]?.title}
      />
    </div>
  )
}
