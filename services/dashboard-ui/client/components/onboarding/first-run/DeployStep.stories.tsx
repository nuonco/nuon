export default {
  title: 'Features / Onboarding / First run / Deploy',
}

import { MockStory } from './mocks'

// Every story runs the mocked flow from this screen, so all buttons work.

export const OwnApp = () => <MockStory start={{ step: 'deploy' }} />
OwnApp.meta = { fullBleed: true }

export const ExampleApp = () => <MockStory start={{ step: 'deploy', path: 'example', cloud: 'gcp' }} />
ExampleApp.meta = { fullBleed: true }

export const WaitingForConfig = () => <MockStory start={{ step: 'deploy', scenario: { deploy: 'waiting' } }} />
WaitingForConfig.meta = { fullBleed: true }

export const SyncFailed = () => <MockStory start={{ step: 'deploy', scenario: { deploy: 'sync-failed' } }} />
SyncFailed.meta = { fullBleed: true }

export const ExampleSyncFailed = () => <MockStory start={{ step: 'deploy', path: 'example', scenario: { deploy: 'sync-failed' } }} />
ExampleSyncFailed.meta = { fullBleed: true }

export const SlowWait = () => <MockStory start={{ step: 'deploy', path: 'example', scenario: { deploy: 'slow' } }} />
SlowWait.meta = { fullBleed: true }

export const MissingInputDefaults = () => <MockStory start={{ step: 'deploy', cloud: 'azure', scenario: { deploy: 'missing-inputs' } }} />
MissingInputDefaults.meta = { fullBleed: true }

export const KitchenSinkAws = () => <MockStory start={{ step: 'deploy', path: 'example' }} />
KitchenSinkAws.meta = { fullBleed: true }

export const KitchenSinkWaitingForConfig = () => <MockStory start={{ step: 'deploy', path: 'example', scenario: { deploy: 'waiting' } }} />
KitchenSinkWaitingForConfig.meta = { fullBleed: true }