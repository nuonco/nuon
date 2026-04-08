export default {
  title: 'Onboarding/V2 Steps/CloudSetupStep',
}

import { CloudSetupStepContainer } from './CloudSetupStepContainer'

const mockProps = {
  onAdvance: () => {},
  onGoBack: () => {},
  isComplete: false,
  sharedData: {
    onboarding: {
      org_id: 'org-1',
      app_id: 'app-1',
      cloud_provider: 'aws',
      example_app_slug: 'eks-simple',
      status_v2: { status: 'active' },
    },
  },
  setSharedData: () => {},
  nextStepTitle: 'Provisioning',
}

export const Default = () => <CloudSetupStepContainer {...mockProps} />

export const NoCloudProvider = () => (
  <CloudSetupStepContainer
    {...mockProps}
    sharedData={{
      onboarding: {
        org_id: 'org-1',
        app_id: 'app-1',
        cloud_provider: null,
        status_v2: { status: 'active' },
      },
    }}
  />
)

export const NoAppId = () => (
  <CloudSetupStepContainer
    {...mockProps}
    sharedData={{
      onboarding: {
        org_id: 'org-1',
        cloud_provider: 'aws',
        status_v2: { status: 'active' },
      },
    }}
  />
)
