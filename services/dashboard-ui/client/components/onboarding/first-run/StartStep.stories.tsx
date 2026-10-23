export default {
  title: 'Features / Onboarding / First run / Start',
}

import { MockStory } from './mocks'

// Every story runs the mocked flow from this screen, so all buttons work.

export const ChoosePath = () => <MockStory start={{ step: 'start', path: 'example' }} />
ChoosePath.meta = { fullBleed: true }

export const OwnAppDisconnected = () => <MockStory start={{ step: 'start' }} />
OwnAppDisconnected.meta = { fullBleed: true }

export const OwnAppConnected = () => <MockStory start={{ step: 'start', scenario: { github: { status: 'connected', owner: 'acme', repoCount: 3 }, appName: 'acme-api', testCloud: 'aws' } }} />
OwnAppConnected.meta = { fullBleed: true }

export const MissingFields = () => <MockStory start={{ step: 'start', scenario: { showErrors: true } }} />
MissingFields.meta = { fullBleed: true }

export const InvalidName = () => <MockStory start={{ step: 'start', scenario: { github: { status: 'connected', owner: 'acme', repoCount: 3 }, appName: 'My App' } }} />
InvalidName.meta = { fullBleed: true }

export const NoMatchingRepo = () => <MockStory start={{ step: 'start', scenario: { github: { status: 'connected', owner: 'acme', repoCount: 3 }, appName: 'acme-api', testCloud: 'gcp', repoError: 'No repo named acme-api in your GitHub connection.' } }} />
NoMatchingRepo.meta = { fullBleed: true }

export const NameTaken = () => <MockStory start={{ step: 'start', scenario: { github: { status: 'connected', owner: 'acme', repoCount: 3 }, appName: 'acme-api', testCloud: 'azure', appNameError: 'An app template with this name exists' } }} />
NameTaken.meta = { fullBleed: true }

export const GithubCallbackFailed = () => <MockStory start={{ step: 'start', scenario: { github: { status: 'disconnected', error: true } } }} />
GithubCallbackFailed.meta = { fullBleed: true }

export const CreatingExample = () => <MockStory start={{ step: 'start', path: 'example', scenario: { creating: 'aws' } }} />
CreatingExample.meta = { fullBleed: true }