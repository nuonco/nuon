export default {
  title: 'Onboarding/V2 Steps/ProvisioningStep',
}

import { ProvisioningStepContainer } from './ProvisioningStepContainer'

const mockProps = {
  onAdvance: () => {},
  onGoBack: () => {},
  isComplete: false,
  sharedData: {
    onboarding: {
      org_id: 'org-1',
      app_id: 'app-1',
      install_id: 'install-1',
      example_app_slug: 'hello-world',
      status_v2: { status: 'processing' },
    },
  },
  setSharedData: () => {},
  nextStepTitle: 'Next steps',
}

export const Default = () => <ProvisioningStepContainer {...mockProps} />
