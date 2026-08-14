import { ModalStory } from '@/components/__stories__/helpers'
import type { TSlackChannelSubscription } from '@/types'
import { ChannelSubscriptionFormModal } from './ChannelSubscriptionForm'

export default { title: 'Slack/ChannelSubscriptionForm' }

const installations = [
  {
    id: 'sli-001',
    team_id: 'T0123456789',
    team_name: 'nuonco',
    status: 'active' as const,
  },
]

const orgLinks = [
  {
    id: 'slo-001',
    team_id: 'T0123456789',
    org_id: 'org-001',
    status: 'verified' as const,
  },
]

const channels = [
  { id: 'C0123', name: 'deploys', is_member: true } as const,
  { id: 'C0456', name: 'approvals', is_member: true } as const,
  { id: 'C0789', name: 'general', is_member: true } as const,
]

const baseSub: TSlackChannelSubscription = {
  id: 'slcs-001',
  org_id: 'org-001',
  org_link_id: 'slo-001',
  team_id: 'T0123456789',
  channel_id: 'C0123',
  channel_name: 'deploys',
  created_at: '2026-04-01T00:00:00Z',
  updated_at: '2026-04-01T00:00:00Z',
  interests: { all_events: true },
} as TSlackChannelSubscription

export const Create = () => (
  <ModalStory>
    <ChannelSubscriptionFormModal
      mode="create"
      installations={installations}
      orgLinks={orgLinks}
      channels={channels}
      selectedInstallationId="sli-001"
      channelsError={null}
      channelSearch=""
      onChannelSearchChange={() => {}}
      hasMoreChannels={false}
      isLoadingFirstChannelsPage={false}
      isFetchingNextChannelsPage={false}
      onLoadMoreChannels={() => {}}
      isPending={false}
      error={null}
      onSelectInstallation={() => {}}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const CreateNoInstallations = () => (
  <ModalStory>
    <ChannelSubscriptionFormModal
      mode="create"
      installations={[]}
      orgLinks={[]}
      channels={[]}
      selectedInstallationId={null}
      channelsError={null}
      channelSearch=""
      onChannelSearchChange={() => {}}
      hasMoreChannels={false}
      isLoadingFirstChannelsPage={false}
      isFetchingNextChannelsPage={false}
      onLoadMoreChannels={() => {}}
      isPending={false}
      error={null}
      onSelectInstallation={() => {}}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const CreateLoadingMore = () => (
  <ModalStory>
    <ChannelSubscriptionFormModal
      mode="create"
      installations={installations}
      orgLinks={orgLinks}
      channels={channels}
      selectedInstallationId="sli-001"
      channelsError={null}
      channelSearch=""
      onChannelSearchChange={() => {}}
      hasMoreChannels
      isLoadingFirstChannelsPage={false}
      isFetchingNextChannelsPage
      onLoadMoreChannels={() => {}}
      isPending={false}
      error={null}
      onSelectInstallation={() => {}}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const EditOrgWide = () => (
  <ModalStory>
    <ChannelSubscriptionFormModal
      mode="edit"
      subscription={baseSub}
      isPending={false}
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const EditInstallScoped = () => (
  <ModalStory>
    <ChannelSubscriptionFormModal
      mode="edit"
      subscription={{
        ...baseSub,
        match: { installs: { ids: ['inst_a', 'inst_b'] } },
      }}
      isPending={false}
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const EditLabelScoped = () => (
  <ModalStory>
    <ChannelSubscriptionFormModal
      mode="edit"
      subscription={{
        ...baseSub,
        match: {
          installs: {
            selector: { match_labels: { env: 'prod', tier: 'critical' } },
          },
        },
      }}
      isPending={false}
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const EditComponentsScoped = () => (
  <ModalStory>
    <ChannelSubscriptionFormModal
      mode="edit"
      subscription={{
        ...baseSub,
        match: { components: { ids: ['cmp_a'] } },
      }}
      isPending={false}
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)
