export default {
  title: 'Features / Onboarding / V2 steps / Next steps step',
}

import { NextStepsStepContainer } from './NextStepsStepContainer'

const mockProps = {
  onAdvance: () => {},
  onGoBack: () => {},
  isComplete: false,
  sharedData: {
    onboarding: {
      org_id: 'org-1',
      install_id: 'install-1',
      status_v2: { status: 'active' },
    },
  },
  setSharedData: () => {},
  nextStepTitle: 'Done',
}

export const Default = () => <NextStepsStepContainer {...mockProps} />
