export default {
  title: 'Datadog/CreateConnection',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateConnectionModal } from './CreateConnection'

export const Default = () => (
  <ModalStory>
    <CreateConnectionModal
      isPending={false}
      error={null}
      onSubmit={() => undefined}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <CreateConnectionModal
      isPending
      error={null}
      onSubmit={() => undefined}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <CreateConnectionModal
      isPending={false}
      error={
        {
          error: 'datadog rejected the api key',
          description:
            'Datadog returned 403 — the key is invalid for the chosen site.',
          status: 400,
        } as any
      }
      onSubmit={() => undefined}
    />
  </ModalStory>
)
