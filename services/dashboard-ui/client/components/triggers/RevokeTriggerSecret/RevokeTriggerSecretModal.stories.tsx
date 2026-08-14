export default {
  title: 'Triggers/RevokeTriggerSecretModal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { RevokeTriggerSecretModal } from './RevokeTriggerSecretModal'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <RevokeTriggerSecretModal onConfirm={noop} />
  </ModalStory>
)
