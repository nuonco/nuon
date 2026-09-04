export default {
  title: 'Installs/EditLabels Mockup',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { LabelsFormMockup } from './LabelsFormMockup'

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

export const ManyLabels = () => (
  <ModalStory label="Open mockup (20 labels)">
    <LabelsFormMockup
      labels={manyLabels}
      defaultLabels={{
        tier: 'gold',
        region: '{{ .nuon.cloud_account.aws.region }}',
      }}
      onSubmit={noop}
    />
  </ModalStory>
)

export const FewLabels = () => (
  <ModalStory label="Open mockup (3 labels)">
    <LabelsFormMockup
      labels={{ env: 'production', team: 'platform', region: 'us-east-1' }}
      defaultLabels={{ tier: 'gold' }}
      onSubmit={noop}
    />
  </ModalStory>
)

export const NoDefaultLabels = () => (
  <ModalStory label="Open mockup (no default labels)">
    <LabelsFormMockup
      labels={{ env: 'staging', team: 'platform' }}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Empty = () => (
  <ModalStory label="Open mockup (empty)">
    <LabelsFormMockup labels={{}} onSubmit={noop} />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory label="Open mockup (saving)">
    <LabelsFormMockup
      labels={{ env: 'staging', team: 'platform' }}
      defaultLabels={{ tier: 'gold' }}
      isPending
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory label="Open mockup (error)">
    <LabelsFormMockup
      labels={{ env: 'staging' }}
      error={{
        error: 'Unable to update labels',
        description: 'Label keys must be unique',
        user_error: true,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
