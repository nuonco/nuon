import { useCallback, useState } from 'react'
import { WizardNav } from '../WizardNav'
import { WizardStepView } from '../WizardStepView'

interface IOnboardingWizardLayout {
  skipHref?: string | null
  onSkip?: () => void
}

export const OnboardingWizardLayout = ({ skipHref, onSkip }: IOnboardingWizardLayout) => {
  const [isScrolled, setIsScrolled] = useState(false)

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    setIsScrolled(e.currentTarget.scrollTop > 0)
  }, [])

  return (
    <div className="h-screen flex flex-col bg-background relative">
      <WizardNav isScrolled={isScrolled} skipHref={skipHref} onSkip={onSkip} />
      <div
        className="flex-1 overflow-y-auto px-6 pt-14 pb-8"
        onScroll={handleScroll}
      >
        <div className="max-w-4xl mx-auto w-full">
          <WizardStepView />
        </div>
      </div>
    </div>
  )
}
