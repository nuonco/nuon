export default {
  title: 'Operations/ActivityPins',
}

import { Button } from '@/components/common/Button'
import { ActivityPins, type TActivityPinCard } from './ActivityPins'

const panel = <Button variant="secondary">Edit pins</Button>

const pins: TActivityPinCard[] = [
  { id: 'rb-1', kind: 'runbook', name: 'Rotate secrets', onRun: () => {} },
  { id: 'act-1', kind: 'action', name: 'Restart workers', onRun: () => {} },
  { id: 'rb-2', kind: 'runbook', name: 'Drain nodes', onRun: () => {} },
]

export const Empty = () => (
  <ActivityPins
    items={[]}
    panel={<Button variant="secondary">Pin shortcuts</Button>}
  />
)

export const WithPins = () => <ActivityPins items={pins} panel={panel} />

export const FourPins = () => (
  <ActivityPins
    items={[
      ...pins,
      { id: 'act-2', kind: 'action', name: 'Flush cache', onRun: () => {} },
    ]}
    panel={panel}
  />
)

export const Loading = () => (
  <ActivityPins
    items={[
      { id: 'rb-1', kind: 'runbook', name: '', loading: true },
      { id: 'act-1', kind: 'action', name: '', loading: true },
    ]}
    panel={panel}
  />
)

export const Unavailable = () => (
  <ActivityPins
    items={[
      {
        id: 'rb-1',
        kind: 'runbook',
        name: 'Rotate secrets',
        missing: true,
        onRemove: () => {},
      },
      { id: 'act-1', kind: 'action', name: 'Restart workers', onRun: () => {} },
    ]}
    panel={panel}
  />
)
