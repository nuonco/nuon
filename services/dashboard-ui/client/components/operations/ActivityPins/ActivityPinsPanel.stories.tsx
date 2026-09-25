export default {
  title: 'Operations/ActivityPinsPanel',
}

import { PanelStory } from '@/components/__stories__/helpers'
import { SearchInput } from '@/components/common/SearchInput'
import { ActivityPinList } from './ActivityPinList'
import { ActivityPinsPanelView } from './ActivityPinsPanel'

const search = (
  <SearchInput
    placeholder="Search by name or ID..."
    value=""
    onChange={() => {}}
  />
)

const runbooks = (
  <ActivityPinList
    title="Runbooks"
    search={search}
    emptyTitle="No runbooks yet"
    emptyMessage="Runbooks show up here once they are defined on the app and synced to this install."
    onToggle={() => {}}
    items={[
      {
        id: 'rb-1',
        name: 'Rotate secrets',
        description: 'Issues new credentials and rolls the API and workers.',
        checked: true,
      },
      {
        id: 'rb-2',
        name: 'Drain nodes',
        description: 'Moves workloads off a node before it is replaced.',
        checked: false,
      },
    ]}
  />
)

const actions = (disabled: boolean) => (
  <ActivityPinList
    title="Actions"
    hint="Only actions with a manual trigger are listed."
    search={search}
    emptyTitle="No actions yet"
    emptyMessage="Actions with a manual trigger show up here once they are on this install."
    onToggle={() => {}}
    items={[
      { id: 'act-1', name: 'Restart workers', checked: true },
      { id: 'act-2', name: 'Flush cache', checked: false, disabled },
    ]}
  />
)

export const Default = () => (
  <PanelStory label="Pin shortcuts">
    <ActivityPinsPanelView
      pinnedCount={2}
      runbooks={runbooks}
      actions={actions(false)}
    />
  </PanelStory>
)

export const AtLimit = () => (
  <PanelStory label="Pin shortcuts">
    <ActivityPinsPanelView
      pinnedCount={4}
      runbooks={runbooks}
      actions={actions(true)}
    />
  </PanelStory>
)
