import type { TSlackChannelSubscription, TSlackOrgLink } from '@/types'
import { ChannelSubscriptionsTable } from './ChannelSubscriptionsTable'

export default { title: 'Slack/ChannelSubscriptionsTable' }

const links: TSlackOrgLink[] = [
  {
    id: 'slo-001',
    team_id: 'T0123456789',
    org_id: 'org-001',
    status: 'verified',
  },
]

const subs: TSlackChannelSubscription[] = [
  {
    id: 'slc-001',
    channel_id: 'C0123456789',
    channel_name: 'deploys',
    team_id: 'T0123456789',
    org_link_id: 'slo-001',
    org_id: 'org-001',
    interests: { all_events: true },
    created_at: '2026-04-30T15:00:00Z',
  },
  {
    id: 'slc-002',
    channel_id: 'C9876543210',
    channel_name: 'approvals',
    team_id: 'T0123456789',
    org_link_id: 'slo-001',
    org_id: 'org-001',
    interests: {
      resources: {
        installs: {
          outcome: 'completion',
          approval_requests: true,
          approval_responses: true,
        },
        components: {
          outcome: 'failures',
          approval_requests: true,
          approval_responses: true,
        },
      },
    },
    created_at: '2026-04-25T15:00:00Z',
  },
  {
    id: 'slc-003',
    channel_id: 'C5555555555',
    channel_name: 'silent',
    team_id: 'T0123456789',
    org_link_id: 'slo-001',
    org_id: 'org-001',
    interests: {},
    created_at: '2026-04-20T15:00:00Z',
  },
  // Per-install scope — describeMatch renders "2 installs".
  {
    id: 'slc-004',
    channel_id: 'C4444444444',
    channel_name: 'install-pinned',
    team_id: 'T0123456789',
    org_link_id: 'slo-001',
    org_id: 'org-001',
    interests: { all_events: true },
    match: { installs: { ids: ['inst_a', 'inst_b'] } },
    created_at: '2026-04-15T15:00:00Z',
  },
  // Components by labels — describeMatch renders "Components: env=prod".
  {
    id: 'slc-005',
    channel_id: 'C3333333333',
    channel_name: 'prod-components',
    team_id: 'T0123456789',
    org_link_id: 'slo-001',
    org_id: 'org-001',
    interests: { all_events: true },
    match: {
      components: { selector: { match_labels: { env: 'prod' } } },
    },
    created_at: '2026-04-10T15:00:00Z',
  },
  // Empty TargetMatch{} — describeMatch renders "Any actions".
  {
    id: 'slc-006',
    channel_id: 'C2222222222',
    channel_name: 'all-actions',
    team_id: 'T0123456789',
    org_link_id: 'slo-001',
    org_id: 'org-001',
    interests: { all_events: true },
    match: { actions: {} },
    created_at: '2026-04-05T15:00:00Z',
  },
]

export const Default = () => (
  <ChannelSubscriptionsTable data={subs} links={links} isLoading={false} />
)
export const Loading = () => (
  <ChannelSubscriptionsTable data={[]} links={[]} isLoading={true} />
)
export const Empty = () => (
  <ChannelSubscriptionsTable data={[]} links={links} isLoading={false} />
)
