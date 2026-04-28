export default {
  title: 'Webhooks/DeleteWebhook',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeleteWebhookModal } from './DeleteWebhook'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <DeleteWebhookModal
      webhookUrl="https://example.com/webhooks/nuon"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <DeleteWebhookModal
      webhookUrl="https://example.com/webhooks/nuon"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <DeleteWebhookModal
      webhookUrl="https://example.com/webhooks/nuon"
      isPending={false}
      error={{
        error: 'webhook not found',
        description: '',
        user_error: true,
        status: 404,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
