export default {
  title: 'Features / Onboarding / First run / Connect',
}

import { MockStory } from './mocks'

// Every story runs the mocked flow from this screen, so all buttons work.

export const WaitingForPush = () => <MockStory start={{ step: 'connect' }} />
WaitingForPush.meta = { fullBleed: true }

export const PushWarning = () => <MockStory start={{ step: 'connect', scenario: { confirmSkip: true } }} />
PushWarning.meta = { fullBleed: true }

export const PushDetected = () => <MockStory start={{ step: 'connect', scenario: { detected: true } }} />
PushDetected.meta = { fullBleed: true }

export const PromptUnavailable = () => <MockStory start={{ step: 'connect', cloud: 'gcp', scenario: { promptFailed: true } }} />
PromptUnavailable.meta = { fullBleed: true }
