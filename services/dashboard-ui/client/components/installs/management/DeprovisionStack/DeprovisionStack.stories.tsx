export default {
  title: 'Installs/DeprovisionStack',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeprovisionStackModal } from './DeprovisionStack'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <DeprovisionStackModal installName="prod-acme" onDismiss={noop} />
  </ModalStory>
)
