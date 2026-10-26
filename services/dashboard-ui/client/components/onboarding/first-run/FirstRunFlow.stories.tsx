export default {
  title: 'Features / Onboarding / First run / Clickable flow',
}

import { MockStory } from './mocks'

// Remembers where you are across reloads; the other stories always open fresh.
export const ClickThrough = () => <MockStory start={{ step: 'intro' }} persist />
ClickThrough.meta = { fullBleed: true }
