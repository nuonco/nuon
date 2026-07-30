export default {
  title: 'InstallHealth/ResetHealthWindow',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ResetHealthWindowModal } from './ResetHealthWindow'

export const Default = () => (
  <ModalStory>
    <ResetHealthWindowModal installId="inl123" />
  </ModalStory>
)
