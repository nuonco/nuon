export default {
  title: 'Webhooks/CreateWebhook',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateWebhookModal } from './CreateWebhook'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <CreateWebhookModal isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <CreateWebhookModal isPending={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const WithConflictError = () => (
  <ModalStory>
    <CreateWebhookModal
      isPending={false}
      error={{
        error:
          'A webhook with this URL already exists for this org. Delete the existing webhook to recreate it.',
        description: '',
        user_error: true,
        status: 409,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithGenericError = () => (
  <ModalStory>
    <CreateWebhookModal
      isPending={false}
      error={{
        error: 'webhook_url must use http or https scheme',
        description: '',
        user_error: true,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithInterestsValidationError = () => (
  <ModalStory>
    <CreateWebhookModal
      isPending={false}
      error={{
        error: 'invalid interests: unknown op "foo" for resource "installs"',
        description: '',
        user_error: true,
        status: 400,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
