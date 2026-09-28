export default {
  title: 'Features / Onboarding / V2 steps / App profile step',
}

import { AppProfileStepContainer } from './AppProfileStepContainer'

const mockProps = {
  onAdvance: () => {},
  onGoBack: () => {},
  isComplete: false,
  sharedData: {
    onboarding: {
      org_id: 'org-1',
      status_v2: { status: 'active' },
    },
  },
  setSharedData: () => {},
  nextStepTitle: 'Cloud setup',
}

export const Default = () => <AppProfileStepContainer {...mockProps} />
