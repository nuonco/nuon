export default {
  title: 'Features / Onboarding / First run / Intro',
}

import { MockStory } from './mocks'

// Every story runs the mocked flow from this screen, so all buttons work.

export const Default = () => <MockStory start={{ step: 'intro' }} />
Default.meta = { fullBleed: true }
