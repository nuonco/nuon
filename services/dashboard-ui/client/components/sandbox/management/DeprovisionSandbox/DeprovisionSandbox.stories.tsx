export default {
  title: 'Sandbox/DeprovisionSandbox',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeprovisionSandboxModal } from './DeprovisionSandbox'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <DeprovisionSandboxModal
      installName="prod-acme"
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
    <DeprovisionSandboxModal
      installName="prod-acme"
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
    <DeprovisionSandboxModal
      installName="prod-acme"
      installId="install-1"
      isPending={false}
      error={{ error: 'Deprovision failed: resources still in use' }}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)
