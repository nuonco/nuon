import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { TAppLabelKey } from '@/lib/ctl-api/apps/get-app-labels'
import { AppLabels } from './AppLabels'

export default {
  title: 'Apps/AppLabels',
}

const mockLabels: TAppLabelKey[] = [
  {
    key: 'service',
    color: '#2563eb',
    default_color: '#2563eb',
    is_override: false,
    values: ['alb', 'coder', 'eks', 'grafana', 'infra', 'kubernetes', 'prometheus', 'rds'],
    entity_types: ['action', 'runbook'],
    usage_count: 25,
  },
  {
    key: 'type',
    color: '#dc2626',
    default_color: '#dc2626',
    is_override: false,
    values: ['break-glass', 'report', 'setup', 'step', 'tool'],
    entity_types: ['action', 'runbook'],
    usage_count: 25,
  },
  {
    key: 'destructive',
    color: '#a21caf',
    default_color: '#16a34a',
    is_override: true,
    values: ['true'],
    entity_types: ['action'],
    usage_count: 1,
  },
]

const noop = () => {}

const ResetAction = () => (
  <Button variant="ghost" onClick={noop}>
    <Icon variant="ArrowCounterClockwiseIcon" size={16} />
    Reset all to defaults
  </Button>
)

export const Default = () => (
  <AppLabels
    labels={mockLabels}
    resetAction={<ResetAction />}
    onOverride={noop}
    onRemoveOverride={noop}
  />
)

export const Loading = () => (
  <AppLabels labels={[]} isLoading onOverride={noop} onRemoveOverride={noop} />
)

export const Empty = () => (
  <AppLabels labels={[]} onOverride={noop} onRemoveOverride={noop} />
)
