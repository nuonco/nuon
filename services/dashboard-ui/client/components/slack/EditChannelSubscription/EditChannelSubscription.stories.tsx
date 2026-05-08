import { ModalStory } from '@/components/__stories__/helpers'
import type { TSlackChannelSubscription } from '@/types'
import { EditChannelSubscriptionModal } from './EditChannelSubscription'

export default { title: 'Slack/EditChannelSubscription' }

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

export const OrgWide = () => (
  <ModalStory>
    <EditChannelSubscriptionModal
      subscription={baseSub}
      isPending={false}
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const InstallScoped = () => (
  <ModalStory>
    <EditChannelSubscriptionModal
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

export const LabelScoped = () => (
  <ModalStory>
    <EditChannelSubscriptionModal
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

export const ComponentsScoped = () => (
  <ModalStory>
    <EditChannelSubscriptionModal
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
