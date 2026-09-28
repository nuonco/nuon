export default {
  title: 'Onboarding/First run/Stack',
}

import { StackStepView, type IStackStepView } from './StackStep'
import { StoryFrame } from './StoryFrame'

// GCP and Azure render the workflow page's stack views below the card in the
// product; those have their own stories under Workflows/StepDetails.
const Harness = (overrides: Partial<IStackStepView>) => {
  const cloud = overrides.cloud ?? 'aws'
  return (
    <StoryFrame step="stack" cloud={cloud}>
      <StackStepView
        cloud={cloud}
        appName="acme-api"
        region={cloud === 'gcp' ? 'us-central1' : cloud === 'azure' ? 'eastus' : 'us-east-1'}
        phase="generating"
        quickLinkUrl="https://us-east-1.console.aws.amazon.com/cloudformation/home"
        onLaunch={() => {}}
        onContinue={() => {}}
        onBack={() => {}}
        {...overrides}
      />
    </StoryFrame>
  )
}

export const AwsGenerating = () => <Harness />
AwsGenerating.meta = { fullBleed: true }

export const AwsReady = () => <Harness phase="ready" />
AwsReady.meta = { fullBleed: true }

export const AwsLaunched = () => <Harness phase="launched" />
AwsLaunched.meta = { fullBleed: true }

export const GcpReady = () => <Harness cloud="gcp" phase="ready" quickLinkUrl={undefined} />
GcpReady.meta = { fullBleed: true }

export const AzureReady = () => <Harness cloud="azure" phase="ready" quickLinkUrl={undefined} />
AzureReady.meta = { fullBleed: true }

export const GenerationFailed = () => (
  <Harness phase="error" errorDescription="Stack template rendering failed: missing input vpc_id." />
)
GenerationFailed.meta = { fullBleed: true }
