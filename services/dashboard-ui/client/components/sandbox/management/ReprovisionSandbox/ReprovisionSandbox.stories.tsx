export default {
  title: 'Sandbox/ReprovisionSandbox',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ReprovisionSandboxModal } from './ReprovisionSandbox'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <ReprovisionSandboxModal
      installId="install-1"
      isPending={false}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ReprovisionSandboxModal
      installId="install-1"
      isPending={true}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ReprovisionSandboxModal
      installId="install-1"
      isPending={false}
      error={{ error: 'Reprovision failed: timeout' }}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)
