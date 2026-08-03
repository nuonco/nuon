export default {
  title: 'Installs/ReprovisionStack',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ReprovisionStackModal } from './ReprovisionStack'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <ReprovisionStackModal
      installId="install-1"
      installName="acme-prod"
      isPending={false}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ReprovisionStackModal
      installId="install-1"
      installName="acme-prod"
      isPending={true}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ReprovisionStackModal
      installId="install-1"
      installName="acme-prod"
      isPending={false}
      error={{ error: 'Reprovision failed: stack version run timed out' }}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)
