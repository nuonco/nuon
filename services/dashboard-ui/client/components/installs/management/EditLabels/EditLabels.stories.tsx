export default {
  title: 'Installs/EditLabels',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { EditLabelsModal } from './EditLabels'

const noop = () => {}

const manyLabels: Record<string, string> = Object.fromEntries([
  ['env', 'production'],
  ['team', 'platform'],
  ['region', 'us-east-1'],
  ['tier', 'gold'],
  ['cost-center', 'eng-4417'],
  ['owner', 'payments'],
  ['channel', 'stable'],
  ['compliance', 'soc2'],
  ['data-residency', 'us'],
  ['support-plan', 'enterprise'],
  ['cluster', 'blue'],
  ['network', 'private'],
  ['backup', 'daily'],
  ['maintenance-window', 'sunday-02-00'],
  ['contract', 'annual'],
  ['onboarding', 'complete'],
  ['sla', '99-9'],
  ['runtime', 'eks'],
  ['release-train', 'weekly'],
  ['pager', 'oncall-platform'],
])

export const Default = () => (
  <ModalStory>
    <EditLabelsModal
      labels={{ env: 'production', team: 'platform', region: 'us-east-1' }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const ManyLabels = () => (
  <ModalStory label="Open modal with 20 labels">
    <EditLabelsModal
      labels={manyLabels}
      defaultLabels={{
        tier: 'gold',
        region: '{{ .nuon.cloud_account.aws.region }}',
      }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithDefaultLabels = () => (
  <ModalStory>
    <EditLabelsModal
      labels={{ env: 'production' }}
      defaultLabels={{
        tier: 'gold',
        region: '{{ .nuon.cloud_account.aws.region }}',
      }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Empty = () => (
  <ModalStory>
    <EditLabelsModal
      labels={{}}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const NoDefaultLabels = () => (
  <ModalStory label="Open modal without default labels">
    <EditLabelsModal
      labels={{ env: 'staging', team: 'platform' }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <EditLabelsModal
      labels={{ env: 'staging' }}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <EditLabelsModal
      labels={{ env: 'staging' }}
      isPending={false}
      error={{
        error: 'Something went wrong',
        description: '',
        user_error: true,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
