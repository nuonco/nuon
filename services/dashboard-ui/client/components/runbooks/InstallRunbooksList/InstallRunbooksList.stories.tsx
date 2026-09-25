export default {
  title: 'Runbooks/InstallRunbooksList',
}

import { Button } from '@/components/common/Button'
import { SearchInput } from '@/components/common/SearchInput'
import { LatestRunbookRunCard } from '@/components/runbooks/LatestRunbookRunCard'
import type { TInstallRunbookRun } from '@/lib/ctl-api/installs/runbooks'
import {
  InstallRunbooksList,
  type TInstallRunbookListItem,
} from './InstallRunbooksList'

const run: TInstallRunbookRun = {
  id: 'rbr-1',
  created_at: '2026-09-22T14:00:00Z',
  updated_at: '2026-09-22T14:08:00Z',
  status: 'success',
  install_workflow_id: 'wf-1',
  install_workflow: {
    id: 'wf-1',
    status: {
      status: 'success',
      status_human_description: 'Runbook finished.',
    },
  },
}

const cardActions = (withReadme = true) => (
  <div className="flex items-center gap-2">
    {withReadme ? (
      <Button variant="secondary" size="sm">
        Readme
      </Button>
    ) : null}
    <Button variant="secondary" size="sm">
      Run runbook
    </Button>
  </div>
)

const latestRun = (entry?: TInstallRunbookRun) => (
  <LatestRunbookRunCard flush run={entry} href={entry ? '#' : undefined} />
)

const item = (
  overrides: Partial<TInstallRunbookListItem>
): TInstallRunbookListItem => ({
  id: 'rbk-1',
  name: 'rotate-secrets',
  href: '#',
  description: 'Rotates API keys and secrets for this install.',
  stepCount: 4,
  actions: cardActions(),
  latestRun: latestRun(run),
  ...overrides,
})

const items: TInstallRunbookListItem[] = [
  item({}),
  item({
    id: 'rbk-2',
    name: 'failover-database',
    description: 'Fails over the primary database to the replica.',
    stepCount: 6,
  }),
  item({
    id: 'rbk-3',
    name: 'drain-nodes',
    description: undefined,
    stepCount: 1,
    actions: cardActions(false),
    latestRun: latestRun({
      ...run,
      install_workflow: {
        id: 'wf-2',
        status: {
          status: 'error',
          status_human_description: 'Runbook failed.',
        },
      },
    }),
  }),
]

export const Default = () => (
  <InstallRunbooksList
    items={items}
    search={
      <SearchInput
        placeholder="Search by name or ID..."
        value=""
        onChange={() => {}}
      />
    }
    pagination={{ hasNext: true, offset: 0, limit: 10 }}
  />
)

export const NeverRun = () => (
  <InstallRunbooksList items={[item({ latestRun: latestRun() })]} />
)

export const Removed = () => (
  <InstallRunbooksList
    items={[
      item({}),
      item({
        id: 'rbk-2',
        name: 'legacy-restore',
        removed: true,
        actions: (
          <Button variant="secondary" size="sm" disabled>
            Run runbook
          </Button>
        ),
      }),
    ]}
  />
)

export const NoResults = () => (
  <InstallRunbooksList
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

export const Empty = () => <InstallRunbooksList items={[]} />

export const Loading = () => <InstallRunbooksList items={[]} loading />
