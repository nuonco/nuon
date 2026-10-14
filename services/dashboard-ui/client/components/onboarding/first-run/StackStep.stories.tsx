export default {
  title: 'Features / Onboarding / First run / Stack',
}

import { StackStepView, type IStackStepView } from './StackStep'
import { StoryFrame } from './StoryFrame'

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

const since = Date.now() - 23_000

export const AwsGenerating = () => (
  <Harness
    progressSince={since}
    progress={[
      { label: 'Create the install', state: 'done' },
      { label: 'Plan the install workflow', state: 'done' },
      { label: 'Generate the install state', state: 'done' },
      { label: "Create the runner's identity", state: 'active' },
      { label: 'Render the CloudFormation template', state: 'pending' },
      { label: 'Publish the CloudFormation stack link', state: 'pending' },
    ]}
  />
)
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
