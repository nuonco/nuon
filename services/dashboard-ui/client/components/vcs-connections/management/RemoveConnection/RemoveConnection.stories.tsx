export default {
  title: 'VCS Connections/RemoveConnection',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { RemoveConnectionModal } from './RemoveConnection'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <RemoveConnectionModal
      connectionName="acme-corp"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <RemoveConnectionModal
      connectionName="acme-corp"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <RemoveConnectionModal
      connectionName="acme-corp"
      isPending={false}
      error={{ error: 'Unable to remove VCS connection' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)
