import { useState, useEffect } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

const PROVISION_STEPS = [
  'Queued',
  'Provisioning infrastructure',
  'Configuring network',
  'Starting services',
  'Active',
]

export const ProvisioningStep = ({ onAdvance, nextStepTitle }: IWizardStepComponentProps) => {
  const [activeIndex, setActiveIndex] = useState(0)

  useEffect(() => {
    if (activeIndex >= PROVISION_STEPS.length - 1) return
    const timer = setTimeout(() => setActiveIndex((i) => i + 1), 1500)
    return () => clearTimeout(timer)
  }, [activeIndex])

  const isDone = activeIndex >= PROVISION_STEPS.length - 1

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-3">
        {PROVISION_STEPS.map((step, idx) => {
          const isComplete = idx < activeIndex
          const isActive = idx === activeIndex

          return (
            <div key={step} className="flex items-center gap-3">
              <div className="w-5 h-5 flex items-center justify-center shrink-0">
                {isComplete ? (
                  <Icon variant="CheckCircle" size={20} weight="fill" />
                ) : isActive ? (
                  <Icon variant="Loading" size={20} />
                ) : (
                  <div className="w-4 h-4 rounded-full border-2" />
                )}
              </div>
              <Text variant="body" theme={!isComplete && !isActive ? 'neutral' : undefined} weight={isActive ? 'strong' : undefined}>
                {step}
              </Text>
            </div>
          )
        })}
      </div>

      <div className="flex self-end">
        <Button type="button" variant="primary" disabled={!isDone} onClick={onAdvance}>
          {nextStepTitle ?? 'Continue'} <Icon variant="CaretRight" weight="bold" />
        </Button>
      </div>
    </div>
  )
}
