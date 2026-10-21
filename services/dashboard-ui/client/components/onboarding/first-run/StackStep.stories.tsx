export default {
  title: 'Features / Onboarding / First run / Stack',
}

import { MockStory } from './mocks'

// Every story runs the mocked flow from this screen, so all buttons work.

export const AwsGenerating = () => <MockStory start={{ step: 'stack' }} />
AwsGenerating.meta = { fullBleed: true }

export const AwsReady = () => <MockStory start={{ step: 'stack', scenario: { stack: 'ready' } }} />
AwsReady.meta = { fullBleed: true }

export const AwsLaunched = () => <MockStory start={{ step: 'stack', scenario: { stack: 'launched' } }} />
AwsLaunched.meta = { fullBleed: true }

export const GcpReady = () => <MockStory start={{ step: 'stack', cloud: 'gcp', scenario: { stack: 'ready' } }} />
GcpReady.meta = { fullBleed: true }

export const AzureReady = () => <MockStory start={{ step: 'stack', cloud: 'azure', scenario: { stack: 'ready' } }} />
AzureReady.meta = { fullBleed: true }

export const GenerationFailed = () => <MockStory start={{ step: 'stack', scenario: { stack: 'generation-failed' } }} />
GenerationFailed.meta = { fullBleed: true }

export const AwsProvisionFailed = () => <MockStory start={{ step: 'stack', scenario: { stack: 'provision-failed' } }} />
AwsProvisionFailed.meta = { fullBleed: true }

export const KitchenSinkAwsGenerating = () => <MockStory start={{ step: 'stack', path: 'example' }} />
KitchenSinkAwsGenerating.meta = { fullBleed: true }

export const KitchenSinkAwsReady = () => <MockStory start={{ step: 'stack', path: 'example', scenario: { stack: 'ready' } }} />
KitchenSinkAwsReady.meta = { fullBleed: true }

export const KitchenSinkAwsLaunched = () => <MockStory start={{ step: 'stack', path: 'example', scenario: { stack: 'launched' } }} />
KitchenSinkAwsLaunched.meta = { fullBleed: true }

export const KitchenSinkGcpReady = () => <MockStory start={{ step: 'stack', path: 'example', cloud: 'gcp', scenario: { stack: 'ready' } }} />
KitchenSinkGcpReady.meta = { fullBleed: true }

export const KitchenSinkAwsProvisionFailed = () => <MockStory start={{ step: 'stack', path: 'example', scenario: { stack: 'provision-failed' } }} />
KitchenSinkAwsProvisionFailed.meta = { fullBleed: true }