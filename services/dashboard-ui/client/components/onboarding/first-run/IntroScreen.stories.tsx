export default {
  title: 'Features / Onboarding / First run / Intro',
}

import { CopyEditor } from '../CopyEditor'
import { IntroScreen } from './IntroScreen'

export const Default = () => (
  <CopyEditor>
    <IntroScreen onStart={() => {}} />
  </CopyEditor>
)
Default.meta = { fullBleed: true }
