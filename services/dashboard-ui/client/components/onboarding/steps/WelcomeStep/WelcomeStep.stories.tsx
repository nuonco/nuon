export default {
  title: 'Features / Onboarding / V1 steps / Welcome step',
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
