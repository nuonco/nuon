export default {
  title: 'Features / Pull requests / List',
}

import type { TPreviewSourcePR } from '@/types'
import { PullRequestList } from './PullRequestList'

const basePath = '/org-1/apps/app-1/branches/br-1'

const pullRequests: TPreviewSourcePR[] = [
  {
    pr_number: 128,
    title: 'Retry checkout when the payment provider times out',
    head_ref: 'feat/checkout-retries',
    head_sha: 'a1b2c3d4e5f67890abcd',
    url: 'https://github.com/acme/payments/pull/128',
    author: 'devon',
    is_draft: true,
    labels: [{ name: 'payments', color: '0e8a16' }],
    created_at: '2026-10-26T15:04:00Z',
    updated_at: '2026-10-27T18:12:00Z',
  },
  {
    pr_number: 127,
    title: 'Show the receipt id on the order page',
    head_ref: 'feat/receipt-id',
    head_sha: 'b2c3d4e5f67890abcdef',
    url: 'https://github.com/acme/payments/pull/127',
    author: 'rory',
    labels: [
      { name: 'payments', color: '0e8a16' },
      { name: 'ui', color: '1d76db' },
    ],
    created_at: '2026-10-25T11:20:00Z',
    updated_at: '2026-10-27T09:40:00Z',
  },
  {
    pr_number: 119,
    title:
      'Align the settlement window with the ledger close, including weekends and the month-end exception',
    head_ref: 'feat/settlement-window',
    head_sha: 'c3d4e5f6789012345678',
    url: 'https://github.com/acme/payments/pull/119',
    author: 'devon',
    labels: [{ name: 'ledger', color: 'fbca04' }],
    created_at: '2026-10-20T08:00:00Z',
    updated_at: '2026-10-24T16:30:00Z',
  },
  {
    pr_number: 104,
    title: 'Pin the runner image digest in the example config',
    head_ref: 'feat/checkout-retries',
    head_sha: 'd4e5f67890123456789a',
    url: 'https://github.com/acme/payments/pull/104',
    author: 'rory',
    created_at: '2026-10-18T13:15:00Z',
    updated_at: '2026-10-18T13:15:00Z',
  },
]

export const Default = () => (
  <PullRequestList pullRequests={pullRequests} basePath={basePath} />
)

export const Empty = () => (
  <PullRequestList pullRequests={[]} basePath={basePath} />
)

export const Loading = () => (
  <PullRequestList pullRequests={[]} basePath={basePath} isLoading />
)
