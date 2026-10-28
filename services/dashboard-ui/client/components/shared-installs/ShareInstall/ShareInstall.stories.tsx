export default {
  title: 'Playground / Customer portal / Share install',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { mockShares } from '../fixtures'
import { ShareInstallModal } from './ShareInstallModal'

const noop = () => {}

export const Default = () => (
  <ModalStory label="Share install">
    <ShareInstallModal
      isPending={false}
      error={null}
      installName="Acme production"
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory label="Share install">
    <ShareInstallModal
      isPending
      error={null}
      installName="Acme production"
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory label="Share install">
    <ShareInstallModal
      isPending={false}
      error={{
        error: 'This install is already shared with that email',
        description: '',
        user_error: true,
      }}
      installName="Acme production"
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithExistingShares = () => (
  <ModalStory label="Share install">
    <ShareInstallModal
      isPending={false}
      error={null}
      installName="Acme production"
      shares={mockShares.filter((share) => share.install_id === 'inst-1')}
      onSubmit={noop}
      onRevoke={noop}
      onResend={noop}
    />
  </ModalStory>
)
