export default {
  title: 'Spotlight/SpotlightModal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { SpotlightModal } from './SpotlightModal'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <SpotlightModal
      orgId="org-1"
      onClose={noop}
      onNavigate={noop}
    />
  </ModalStory>
)

export const WithResults = () => (
  <ModalStory label="Open with org features">
    <SpotlightModal
      orgId="org-1"
      onClose={noop}
      onNavigate={noop}
      orgFeatures={{ 'org-dashboard': true }}
    />
  </ModalStory>
)
