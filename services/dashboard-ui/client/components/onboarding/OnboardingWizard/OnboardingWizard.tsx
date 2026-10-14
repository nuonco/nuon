import { useCallback, useState } from 'react'
import { cn } from '@/utils/classnames'
import { WizardNav } from '../WizardNav'
import { WizardStepView } from '../WizardStepView'

interface IOnboardingWizardLayout {
  skipHref?: string | null
  onSkip?: () => void
  variant?: 'screen' | 'pane'
}

export const OnboardingWizardLayout = ({
  skipHref,
  onSkip,
  variant = 'screen',
}: IOnboardingWizardLayout) => {
  const [isScrolled, setIsScrolled] = useState(false)

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    setIsScrolled(e.currentTarget.scrollTop > 0)
  }, [])

  return (
    <div
      className={cn(
        'flex flex-col bg-background relative min-h-0',
        variant === 'screen' ? 'h-screen' : 'h-full overflow-hidden'
      )}
    >
      <WizardNav
        isScrolled={isScrolled}
        showHeader={variant === 'screen'}
        skipHref={skipHref}
        onSkip={onSkip}
      />
      <div
        className="flex-1 min-h-0 overflow-y-auto px-6 pt-14 pb-8"
        onScroll={handleScroll}
      >
        <div className="max-w-4xl mx-auto w-full">
          <WizardStepView />
        </div>
      </div>
    </div>
  )
}
