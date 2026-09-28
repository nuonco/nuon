export default {
  title: 'Onboarding/First run/Provision',
}

import { ProvisionStepView } from './ProvisionStep'
import { StoryFrame } from './StoryFrame'

export const Default = () => (
  <StoryFrame step="provision">
    <ProvisionStepView
      cloud="aws"
      appName="acme-api"
      region="us-east-1"
      installId="inlk3x9q2m7v4w8p1z6r5t0y2c"
      onFinish={() => {}}
      onBack={() => {}}
    />
  </StoryFrame>
)
Default.meta = { fullBleed: true }

export const Gcp = () => (
  <StoryFrame step="provision" path="example" cloud="gcp">
    <ProvisionStepView
      cloud="gcp"
      appName="Kitchen Sink"
      region="us-central1"
      installId="inlk3x9q2m7v4w8p1z6r5t0y2c"
      onFinish={() => {}}
      onBack={() => {}}
    />
  </StoryFrame>
)
Gcp.meta = { fullBleed: true }
