export default {
  title: 'Webhooks/WebhookForm',
}

import { ModalStory } from '@/components/__stories__/helpers'
import type { TWebhook } from '@/types'
import { WebhookFormModal } from './WebhookForm'

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

export const Create = () => (
  <ModalStory>
    <WebhookFormModal
      mode="create"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const CreatePending = () => (
  <ModalStory>
    <WebhookFormModal
      mode="create"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const CreateWithConflictError = () => (
  <ModalStory>
    <WebhookFormModal
      mode="create"
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

export const CreateWithInterestsValidationError = () => (
  <ModalStory>
    <WebhookFormModal
      mode="create"
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

export const Edit = () => (
  <ModalStory>
    <WebhookFormModal
      mode="edit"
      webhook={baseWebhook}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const EditNoSecretConfigured = () => (
  <ModalStory>
    <WebhookFormModal
      mode="edit"
      webhook={{ ...baseWebhook, has_secret: false }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const EditScopedToInstallLabels = () => (
  <ModalStory>
    <WebhookFormModal
      mode="edit"
      webhook={{
        ...baseWebhook,
        match: {
          installs: { selector: { match_labels: { env: 'prod', tier: '*' } } },
        },
      }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const EditScopedToSpecificInstalls = () => (
  <ModalStory>
    <WebhookFormModal
      mode="edit"
      webhook={{
        ...baseWebhook,
        match: { installs: { ids: ['ins_001', 'ins_002'] } },
      }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const EditPerResourceInterests = () => (
  <ModalStory>
    <WebhookFormModal
      mode="edit"
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

export const EditPending = () => (
  <ModalStory>
    <WebhookFormModal
      mode="edit"
      webhook={baseWebhook}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const EditWithError = () => (
  <ModalStory>
    <WebhookFormModal
      mode="edit"
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
