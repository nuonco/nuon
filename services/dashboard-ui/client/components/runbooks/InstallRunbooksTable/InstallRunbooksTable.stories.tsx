export default {
  title: 'Runbooks/InstallRunbooksTable',
}

import { Text } from '@/components/common/Text'
import { InstallRunbooksTable, type TInstallRunbookRow } from './InstallRunbooksTable'

const mockRows: TInstallRunbookRow[] = Array.from({ length: 3 }, (_, i) => ({
  runbookId: `runbook-${i + 1}`,
  runbookName: `rotate-secrets-${i + 1}`,
  description: (
    <Text variant="subtext" theme="neutral">
      Rotates API keys and credentials.
    </Text>
  ),
  labels: <Text variant="subtext" theme="neutral">production</Text>,
  lastUpdated: <Text variant="subtext" theme="neutral">3 days ago</Text>,
  lastRun: <Text variant="subtext" theme="neutral">2 hours ago</Text>,
  href: `/org-1/installs/install-1/runbooks/runbook-${i + 1}`,
  latestRunHref: `/org-1/installs/install-1/workflows/wf-${i + 1}`,
  installRunbook: { id: `ir-${i + 1}`, runbook_id: `runbook-${i + 1}` } as any,
}))

export const Default = () => (
  <InstallRunbooksTable
    data={mockRows}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <InstallRunbooksTable
    data={[]}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => (
  <InstallRunbooksTable
    data={[]}
    isLoading
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)
