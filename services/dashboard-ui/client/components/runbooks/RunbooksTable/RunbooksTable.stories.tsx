export default {
  title: 'Runbooks/RunbooksTable',
}

import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { RunbooksTable, type TRunbookRow } from './RunbooksTable'

const mockRows: TRunbookRow[] = Array.from({ length: 3 }, (_, i) => ({
  runbookId: `runbook-${i + 1}`,
  runbookName: `rotate-secrets-${i + 1}`,
  description: (
    <Text variant="subtext" theme="neutral">
      Rotates API keys and secrets for the install.
    </Text>
  ),
  labels: (
    <Text variant="subtext" theme="neutral">
      production
    </Text>
  ),
  lastUpdated: (
    <Text variant="subtext" theme="neutral">
      3 days ago
    </Text>
  ),
  href: `/org-1/apps/app-1/runbooks/runbook-${i + 1}`,
}))

export const Default = () => (
  <RunbooksTable
    data={mockRows}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <RunbooksTable
    data={[]}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => (
  <RunbooksTable
    data={[]}
    isLoading
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const InstallScope = () => (
  <RunbooksTable
    scope="install"
    data={mockRows.map((row) => ({
      ...row,
      lastRun: (
        <Text variant="subtext" theme="neutral">
          2 hours ago
        </Text>
      ),
      actions: <Button variant="ghost">Manage</Button>,
    }))}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const InstallScopeRemoved = () => (
  <RunbooksTable
    scope="install"
    data={mockRows.map((row, i) => ({ ...row, removed: i === 0 }))}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)
