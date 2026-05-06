export default {
  title: 'Webhooks/EditWebhook',
}

import { ModalStory } from '@/components/__stories__/helpers'
import type { TWebhook } from '@/types'
import { EditWebhookModal } from './EditWebhook'

const noop = () => {}

const baseWebhook: TWebhook = {
  id: 'whk_001',
  org_id: 'org_001',
  webhook_url: 'https://example.com/webhooks/nuon',
  has_secret: true,
  created_at: '2026-04-01T00:00:00Z',
  updated_at: '2026-04-01T00:00:00Z',
  interests: { all_events: true },
}

export const Default = () => (
  <ModalStory>
    <EditWebhookModal
      webhook={baseWebhook}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const NoSecretConfigured = () => (
  <ModalStory>
    <EditWebhookModal
      webhook={{ ...baseWebhook, has_secret: false }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const PerResourceInterests = () => (
  <ModalStory>
    <EditWebhookModal
      webhook={{
        ...baseWebhook,
        interests: {
          resources: {
            installs: {
              outcome: 'completion',
              approval_requests: true,
              approval_responses: true,
            },
            components: {
              ops: ['deploy'],
              outcome: 'failures',
            },
          },
        },
      }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <EditWebhookModal
      webhook={baseWebhook}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <EditWebhookModal
      webhook={baseWebhook}
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
