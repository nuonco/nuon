import { Card } from '@/components/common/Card'
import type { TAppLabelKey } from '@/lib/ctl-api/apps/get-app-labels'
import { LabelRow } from './LabelRow'

export default {
  title: 'Apps/AppLabels/LabelRow',
}

const noop = () => {}

const base: TAppLabelKey = {
  key: 'service',
  color: '#2563eb',
  default_color: '#2563eb',
  is_override: false,
  values: ['alb', 'eks', 'rds'],
  entity_types: ['action', 'runbook'],
  usage_count: 25,
}

export const Default = () => (
  <Card className="flex flex-col divide-y">
    <LabelRow label={base} onOverride={noop} onRemoveOverride={noop} />
  </Card>
)

export const Overridden = () => (
  <Card className="flex flex-col divide-y">
    <LabelRow
      label={{ ...base, key: 'destructive', color: '#a21caf', default_color: '#16a34a', is_override: true, values: ['true'], entity_types: ['action'], usage_count: 1 }}
      onOverride={noop}
      onRemoveOverride={noop}
    />
  </Card>
)
