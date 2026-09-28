export default {
  title: 'Features / Triggers / Revoke trigger secret modal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { RevokeTriggerSecretModal } from './RevokeTriggerSecretModal'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <RevokeTriggerSecretModal onConfirm={noop} />
  </ModalStory>
)
