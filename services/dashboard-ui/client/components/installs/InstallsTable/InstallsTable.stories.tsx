export default {
  title: 'Features / Installs / Installs table',
}

import { useState } from 'react'
import { Button } from '@/components/common/Button'
import {
  HealthCoverage,
  InstallHealthFilters,
  InstallsTable,
  type InstallRow,
} from './InstallsTable'

const mockRows: InstallRow[] = Array.from({ length: 5 }, (_, i) => ({
  name: `prod-acme-${i + 1}`,
  nameHref: `/org-1/installs/install-${i + 1}`,
  installId: `install-${i + 1}`,
  appName: 'My BYOC App',
  appHref: `/org-1/apps/app-1`,
  statuses: <span className="text-sm text-foreground-muted">active</span>,
  region: <span className="text-sm text-foreground-muted">us-west-2</span>,
  platform: <span className="text-sm text-foreground-muted">AWS</span>,
  labels:
    i % 2 === 0 ? <span className="text-xs font-mono">env: prod</span> : null,
  branch: null,
  group: <span className="text-sm font-mono">production</span>,
  activity: <span className="text-sm text-foreground-muted">{i + 1}h ago</span>,
  updatedAt: new Date(Date.now() - (i + 1) * 60 * 60 * 1000).toISOString(),
  action: (
    <Button size="sm" variant="ghost">
      Manage
    </Button>
  ),
}))

const coverageFixtures: Record<string, string>[] = [
  { 'comp-api': 'healthy', 'comp-db': 'healthy', 'comp-queue': 'healthy' },
  { 'comp-api': 'healthy', 'comp-db': '', 'comp-dns': 'not-applicable' },
  { 'comp-api': 'unknown', 'comp-db': 'healthy', 'comp-cache': 'unset' },
  { 'comp-api': 'degraded', 'comp-db': 'healthy' },
  { 'comp-api': 'healthy', 'comp-db': 'custom-verdict' },
]

const filteredRows: InstallRow[] = mockRows.slice(0, 3).map((row, i) => ({
  ...row,
  healthCoverage: (
    <HealthCoverage componentHealthStatuses={coverageFixtures[i]} />
  ),
}))

const HealthFilters = ({
  initialHealth = [],
  initialLifecyclePhase = [],
}: {
  initialHealth?: string[]
  initialLifecyclePhase?: string[]
}) => {
  const [health, setHealth] = useState(new Set(initialHealth))
  const [lifecyclePhase, setLifecyclePhase] = useState(
    new Set(initialLifecyclePhase)
  )
  return (
    <div className="flex items-center gap-3">
      <InstallHealthFilters
        health={health}
        lifecyclePhase={lifecyclePhase}
        onHealthChange={setHealth}
        onLifecyclePhaseChange={setLifecyclePhase}
      />
    </div>
  )
}

export const Default = () => (
  <InstallsTable
    data={mockRows}
    isLoading={false}
    emptyStateAction={<Button>Create install</Button>}
    filterActions={<HealthFilters />}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const FilteredByHealth = () => (
  <InstallsTable
    data={filteredRows}
    isLoading={false}
    scope="app"
    hasActiveFilters
    showHealthCoverage
    filterActions={
      <HealthFilters
        initialHealth={['healthy', 'unknown']}
        initialLifecyclePhase={['provisioned']}
      />
    }
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const FilteredEmpty = () => (
  <InstallsTable
    data={[]}
    isLoading={false}
    scope="app"
    hasActiveFilters
    showHealthCoverage
    emptyStateAction={<Button>Create install</Button>}
    filterActions={<HealthFilters initialHealth={['unhealthy']} />}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const LoadError = () => (
  <InstallsTable
    data={mockRows}
    isLoading={false}
    scope="app"
    hasActiveFilters
    errorMessage="invalid health value: healthyish"
    emptyStateAction={<Button>Create install</Button>}
    filterActions={<HealthFilters initialHealth={['healthyish']} />}
    pagination={{ hasNext: true, offset: 20, limit: 20 }}
  />
)

export const Empty = () => (
  <InstallsTable
    data={[]}
    isLoading={false}
    emptyStateAction={<Button>Create install</Button>}
    filterActions={<HealthFilters />}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => (
  <InstallsTable
    data={[]}
    isLoading
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const AppScope = () => (
  <InstallsTable
    data={mockRows}
    isLoading={false}
    scope="app"
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const BranchScope = () => (
  <InstallsTable
    data={mockRows}
    isLoading={false}
    scope="branch"
    emptyTitle="No installs in this deployment plan"
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)
