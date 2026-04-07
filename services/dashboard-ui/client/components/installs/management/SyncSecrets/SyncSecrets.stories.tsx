export default {
  title: 'Installs/SyncSecrets',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { SyncSecretsModal } from './SyncSecrets'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <SyncSecretsModal installName="prod-acme" isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <SyncSecretsModal installName="prod-acme" isPending={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <SyncSecretsModal installName="prod-acme" isPending={false} error={{ error: 'Something went wrong' }} onSubmit={noop} />
  </ModalStory>
)
