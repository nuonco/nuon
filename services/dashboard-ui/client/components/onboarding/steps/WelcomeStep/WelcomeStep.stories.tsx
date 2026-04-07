export default {
  title: 'Onboarding/V1 Steps/WelcomeStep',
}

import { WelcomeStep } from './WelcomeStep'

export const Default = () => (
  <WelcomeStep
    isPending={false}
    nextStepTitle="Continue"
    onSubmit={(e) => e.preventDefault()}
    onAdvance={() => {}}
  />
)

export const Submitting = () => (
  <WelcomeStep
    isPending={true}
    onSubmit={(e) => e.preventDefault()}
    onAdvance={() => {}}
  />
)
