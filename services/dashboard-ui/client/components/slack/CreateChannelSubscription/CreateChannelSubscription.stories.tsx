import { ModalStory } from '@/components/__stories__/helpers'
import { CreateChannelSubscriptionModal } from './CreateChannelSubscription'

export default { title: 'Slack/CreateChannelSubscription' }

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

export const Default = () => (
  <ModalStory>
    <CreateChannelSubscriptionModal
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

export const NoInstallations = () => (
  <ModalStory>
    <CreateChannelSubscriptionModal
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

// Stories don't drive the internal MatchPicker via a prop — the picker
// owns its own state — so this variant doesn't preconfigure scope at the
// modal level. The pre-scoped flow is exercised by EditChannelSubscription
// stories instead, which seed `subscription.match` directly.
export const LoadingMore = () => (
  <ModalStory>
    <CreateChannelSubscriptionModal
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
