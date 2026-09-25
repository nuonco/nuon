export default {
  title: 'Operations/ActivityPinList',
}

import { SearchInput } from '@/components/common/SearchInput'
import {
  ActivityPinList,
  type TActivityPinListItem,
} from './ActivityPinList'

const search = (
  <SearchInput
    placeholder="Search by name or ID..."
    value=""
    onChange={() => {}}
  />
)

const runbooks: TActivityPinListItem[] = [
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
  {
    id: 'rb-3',
    name: 'Failover database',
    description: 'Promotes the replica and repoints the application.',
    checked: false,
  },
]

export const Runbooks = () => (
  <ActivityPinList
    title="Runbooks"
    items={runbooks}
    search={search}
    emptyTitle="No runbooks yet"
    emptyMessage="Runbooks show up here once they are defined on the app and synced to this install."
    onToggle={() => {}}
  />
)

export const ActionsAtLimit = () => (
  <ActivityPinList
    title="Actions"
    hint="Only actions with a manual trigger are listed."
    items={[
      { id: 'act-1', name: 'Restart workers', checked: true },
      {
        id: 'act-2',
        name: 'Flush cache',
        checked: false,
        disabled: true,
      },
      {
        id: 'act-3',
        name: 'Rebuild index',
        checked: false,
        disabled: true,
      },
    ]}
    search={search}
    emptyTitle="No actions yet"
    emptyMessage="Actions with a manual trigger show up here once they are on this install."
    onToggle={() => {}}
  />
)

export const Paginated = () => (
  <ActivityPinList
    title="Runbooks"
    items={runbooks}
    search={search}
    emptyTitle="No runbooks yet"
    emptyMessage="Runbooks show up here once they are defined on the app and synced to this install."
    pagination={{ offset: 8, limit: 8, hasNext: true }}
    onOffsetChange={() => {}}
    onToggle={() => {}}
  />
)

export const NoResults = () => (
  <ActivityPinList
    title="Runbooks"
    items={[]}
    filtered
    search={
      <SearchInput
        placeholder="Search by name or ID..."
        value="missing"
        onChange={() => {}}
      />
    }
    emptyTitle="No runbooks found"
    emptyMessage="Try a different name or ID."
    onToggle={() => {}}
  />
)

export const Empty = () => (
  <ActivityPinList
    title="Actions"
    hint="Only actions with a manual trigger are listed."
    items={[]}
    search={search}
    emptyTitle="No actions yet"
    emptyMessage="Actions with a manual trigger show up here once they are on this install."
    onToggle={() => {}}
  />
)

export const Loading = () => (
  <ActivityPinList
    title="Runbooks"
    items={[]}
    loading
    search={search}
    emptyTitle="No runbooks yet"
    emptyMessage="Runbooks show up here once they are defined on the app and synced to this install."
    onToggle={() => {}}
  />
)
