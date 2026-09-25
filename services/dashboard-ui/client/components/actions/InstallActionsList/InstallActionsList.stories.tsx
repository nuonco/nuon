export default {
  title: 'Actions/InstallActionsList',
}

import { Button } from '@/components/common/Button'
import { SearchInput } from '@/components/common/SearchInput'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { LatestActionRunCard } from '@/components/actions/LatestActionRunCard'
import type { TInstallActionRun } from '@/types'
import {
  InstallActionsList,
  type TInstallActionListItem,
} from './InstallActionsList'

const run = {
  id: 'arun-1',
  created_at: '2026-09-22T14:00:00Z',
  execution_time: 90_000_000_000,
  triggered_by_type: 'manual',
  status_v2: {
    status: 'success',
    status_human_description: 'Action finished.',
  },
} as TInstallActionRun

const runButton = (
  <Button variant="secondary" size="sm">
    Run action
  </Button>
)

const latestRun = (entry?: TInstallActionRun) => (
  <LatestActionRunCard
    flush
    run={entry}
    href={entry ? '#' : undefined}
    trigger={entry ? <ActionTriggerType triggerType="manual" /> : null}
  />
)

const item = (
  overrides: Partial<TInstallActionListItem>
): TInstallActionListItem => ({
  id: 'act-1',
  name: 'rotate-keys',
  href: '#',
  actions: runButton,
  latestRun: latestRun(run),
  ...overrides,
})

const items: TInstallActionListItem[] = [
  item({}),
  item({ id: 'act-2', name: 'health-check' }),
  item({
    id: 'act-3',
    name: 'notify-slack',
    latestRun: latestRun({
      ...run,
      status_v2: { status: 'error', status_human_description: 'Action failed.' },
    } as TInstallActionRun),
  }),
]

export const Default = () => (
  <InstallActionsList
    items={items}
    actions={<Button variant="secondary">Run adhoc action</Button>}
    search={
      <SearchInput
        placeholder="Search by name or ID..."
        value=""
        onChange={() => {}}
      />
    }
    filterActions={<Button variant="secondary">Filter (3)</Button>}
    pagination={{ hasNext: true, offset: 0, limit: 10 }}
  />
)

export const NeverRun = () => (
  <InstallActionsList
    items={[item({ latestRun: latestRun() })]}
    actions={<Button variant="secondary">Run adhoc action</Button>}
  />
)

export const Removed = () => (
  <InstallActionsList
    items={[
      item({}),
      item({
        id: 'act-2',
        name: 'legacy-cleanup',
        removed: true,
        actions: (
          <Button variant="secondary" size="sm" disabled>
            Run action
          </Button>
        ),
      }),
    ]}
  />
)

export const CronPaused = () => (
  <InstallActionsList
    items={items}
    banner={
      <Banner theme="warn">
        <div className="flex flex-col gap-0.5">
          <Text weight="strong">Action crons are paused</Text>
          <Text variant="subtext">
            The runner is offline, so scheduled action runs are disabled. They
            resume when the runner is back online.
          </Text>
        </div>
      </Banner>
    }
  />
)

export const NoResults = () => (
  <InstallActionsList
    items={[]}
    filtered
    search={
      <SearchInput
        placeholder="Search by name or ID..."
        value="rotate"
        onChange={() => {}}
      />
    }
  />
)

export const Empty = () => <InstallActionsList items={[]} />

export const Loading = () => <InstallActionsList items={[]} loading />
