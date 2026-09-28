export default {
  title: 'Features / Installs / Create install / Create install modal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateInstallModal } from './CreateInstallModal'

export const Default = () => (
  <ModalStory label="Open create install modal">
    <CreateInstallModal />
  </ModalStory>
)
