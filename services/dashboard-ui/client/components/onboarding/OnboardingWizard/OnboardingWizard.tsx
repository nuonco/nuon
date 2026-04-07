import { useCallback, useState } from 'react'
import { Button } from '@/components/common/Button'
import { WizardNav } from '../WizardNav'
import { WizardStepView } from '../WizardStepView'

interface IOnboardingWizardLayout {
  onboardingV2: boolean
  skipHref: string | null
}

function SkipOnboardingButton({ href }: { href: string | null }) {
  if (!href) return null

  return (
    <Button
      variant="ghost"
      size="sm"
      href={href}
      className="absolute top-8 right-6 z-20"
    >
      Skip onboarding
    </Button>
  )
}

export const OnboardingWizardLayout = ({
  onboardingV2,
  skipHref,
}: IOnboardingWizardLayout) => {
  const [isScrolled, setIsScrolled] = useState(false)

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    setIsScrolled(e.currentTarget.scrollTop > 0)
  }, [])

  return (
    <div className="h-screen flex flex-col bg-background relative">
      {!onboardingV2 && <SkipOnboardingButton href={skipHref} />}
      <WizardNav isScrolled={isScrolled} />
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
