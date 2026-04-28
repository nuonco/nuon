export default {
  title: 'Webhooks/WebhooksTable',
}

import { WebhooksTable, WebhooksTableSkeleton } from './WebhooksTable'
import type { TWebhook } from '@/types'

const mockWebhooks: TWebhook[] = [
  {
    id: 'wh-1',
    org_id: 'org-1',
    webhook_url: 'https://example.com/webhooks/nuon',
    has_secret: true,
    created_by_id: 'acct-1',
    created_at: '2026-04-20T12:00:00Z',
    updated_at: '2026-04-20T12:00:00Z',
  },
  {
    id: 'wh-2',
    org_id: 'org-1',
    webhook_url: 'https://hooks.acme.io/incoming',
    has_secret: false,
    created_by_id: 'acct-1',
    created_at: '2026-04-22T15:30:00Z',
    updated_at: '2026-04-22T15:30:00Z',
  },
]

export const Default = () => (
  <WebhooksTable data={mockWebhooks} isLoading={false} />
)

export const Empty = () => <WebhooksTable data={[]} isLoading={false} />

export const Loading = () => <WebhooksTableSkeleton />
