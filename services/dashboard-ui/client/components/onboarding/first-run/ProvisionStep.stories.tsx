export default {
  title: 'Features / Onboarding / First run / Provision',
}

import { MockStory } from './mocks'

// Every story runs the mocked flow from this screen, so all buttons work.

export const Default = () => <MockStory start={{ step: 'provision' }} />
Default.meta = { fullBleed: true }

export const Gcp = () => <MockStory start={{ step: 'provision', path: 'example', cloud: 'gcp' }} />
Gcp.meta = { fullBleed: true }

export const KitchenSinkAws = () => <MockStory start={{ step: 'provision', path: 'example' }} />
KitchenSinkAws.meta = { fullBleed: true }