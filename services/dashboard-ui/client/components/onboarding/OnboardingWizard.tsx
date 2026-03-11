import { useCallback, useState } from 'react'
import { Button } from '@/components/common/Button'
import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import { OnboardingWizardProvider, type IOnboardingWizardProps } from '@/providers/onboarding-wizard-provider'
import { WizardNav } from './WizardNav'
import { WizardStepView } from './WizardStepView'

function WizardLayout() {
  const { isStepComplete, getStepMetadata } = useOnboardingJourney()
  const orgCreated = isStepComplete('org_created')
  const orgId = getStepMetadata('org_created', 'org_id') as string | undefined
  const [isScrolled, setIsScrolled] = useState(false)

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    setIsScrolled(e.currentTarget.scrollTop > 0)
  }, [])

  return (
    <div className="h-screen flex flex-col bg-background relative">
      {orgCreated && orgId && (
        <Button
          variant="ghost"
          size="sm"
          href={`/${orgId}/apps`}
          className="absolute top-8 right-6 z-20"
        >
          Skip onboarding
        </Button>
      )}
      <WizardNav isScrolled={isScrolled} />
      <div className="flex-1 overflow-y-auto px-6 pt-14 pb-8" onScroll={handleScroll}>
        <div className="max-w-2xl mx-auto w-full">
          <WizardStepView />
        </div>
      </div>
    </div>
  )
}

export function OnboardingWizard(props: IOnboardingWizardProps) {
  return (
    <OnboardingWizardProvider {...props}>
      <WizardLayout />
    </OnboardingWizardProvider>
  )
}
