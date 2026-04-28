import type { ColumnDef } from '@tanstack/react-table'
import type { TSandboxRunner, TAdminSandboxConfig, TSandboxJob, TSandboxTemplates } from '@/types'
import { Text } from '@/components/common/Text'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Table } from '@/components/common/Table'
import { AdminSection } from '../../shared/AdminSection'
import { AdminMetadataPanel, AdminInfoCard } from '../../shared/AdminMetadata'
import { SandboxConfigPanel } from '../../sandbox/SandboxConfigPanel'

const jobColumns: ColumnDef<TSandboxJob>[] = [
  {
    accessorKey: 'id',
    header: 'Job ID',
    cell: ({ getValue }) => (
      <Text variant="subtext" className="font-mono text-xs">
        {String(getValue())}
      </Text>
    ),
  },
  {
    accessorKey: 'job_type',
    header: 'Job type',
    cell: ({ getValue }) => (
      <Text variant="subtext" className="font-mono text-xs">
        {String(getValue())}
      </Text>
    ),
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ getValue }) => {
      const status = String(getValue())
      const theme =
        status === 'success'
          ? 'success'
          : status === 'failure' || status === 'error'
          ? 'error'
          : status === 'running'
          ? 'info'
          : 'neutral'
      return (
        <Badge theme={theme} size="sm">
          {status}
        </Badge>
      )
    },
  },
  {
    accessorKey: 'created_at',
    header: 'Created',
    cell: ({ getValue }) => (
      <Text variant="subtext" className="text-xs">
        {new Date(String(getValue())).toLocaleString()}
      </Text>
    ),
  },
]

interface IRunnerCard {
  runner: TSandboxRunner
  isSelected: boolean
  onSelect: (runnerId: string) => void
}

const RunnerSelectCard = ({ runner, isSelected, onSelect }: IRunnerCard) => {
  const orgName = runner.runner_group?.org?.name ?? runner.runner_group?.org_id ?? 'Unknown org'

  return (
    <button
      onClick={() => onSelect(runner.id)}
      className={`
        flex flex-col gap-1 p-3 rounded-lg border text-left transition-colors w-full
        ${
          isSelected
            ? 'border-blue-500 bg-blue-50 dark:border-blue-600 dark:bg-blue-950/30'
            : 'border-gray-200 hover:border-gray-300 dark:border-dark-grey-600 dark:hover:border-dark-grey-500'
        }
      `}
    >
      <div className="flex items-center justify-between gap-2">
        <Text variant="subtext" className="font-mono text-xs truncate max-w-[180px]">
          {runner.id}
        </Text>
        <Badge theme={runner.connected ? 'success' : 'error'} size="sm">
          {runner.connected ? 'connected' : 'disconnected'}
        </Badge>
      </div>
      <Text variant="subtext" className="text-xs text-gray-500 dark:text-gray-400">
        {orgName}
      </Text>
    </button>
  )
}

export interface IAdminSandboxSection {
  runners: TSandboxRunner[]
  runnersLoading: boolean
  selectedRunnerId: string | null
  onSelectRunner: (runnerId: string) => void
  configs: TAdminSandboxConfig[]
  configsLoading: boolean
  jobs: TSandboxJob[]
  jobsLoading: boolean
  templates: TSandboxTemplates
  adminEmail: string
  onRefreshConfigs: () => void
}

export const AdminSandboxSection = ({
  runners,
  runnersLoading,
  selectedRunnerId,
  onSelectRunner,
  configs,
  configsLoading,
  jobs,
  jobsLoading,
  templates,
  adminEmail,
  onRefreshConfigs,
}: IAdminSandboxSection) => {
  const selectedRunner = runners.find((r) => r.id === selectedRunnerId)

  const metadata = selectedRunner ? (
    <AdminMetadataPanel>
      <div className="flex flex-col gap-4">
        <AdminInfoCard title="Runner ID" value={selectedRunner.id} copyable />
        <AdminInfoCard title="Runner group ID" value={selectedRunner.runner_group_id} copyable />
        <AdminInfoCard
          title="Org"
          value={selectedRunner.runner_group?.org?.name ?? selectedRunner.runner_group?.org_id}
        />
        <AdminInfoCard title="Status" value={selectedRunner.status} />
        {selectedRunner.status_description && (
          <AdminInfoCard title="Status description" value={selectedRunner.status_description} />
        )}
      </div>
    </AdminMetadataPanel>
  ) : undefined

  return (
    <AdminSection
      title="Sandbox runner controls"
      subtitle="Manage sandbox runners and configure per-job-type behavior"
      metadata={metadata}
    >
      <Card>
        <Text variant="base" weight="strong">
          Connected sandbox runners
        </Text>
        {runnersLoading ? (
          <div className="grid grid-cols-1 gap-2 md:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div
                key={i}
                className="h-14 rounded-lg border border-gray-200 dark:border-dark-grey-600 animate-pulse bg-gray-100 dark:bg-dark-grey-700"
              />
            ))}
          </div>
        ) : runners.length === 0 ? (
          <Text variant="subtext" className="italic">
            No sandbox runners connected
          </Text>
        ) : (
          <div className="grid grid-cols-1 gap-2 md:grid-cols-2 lg:grid-cols-3">
            {runners.map((runner, i) => (
              <RunnerSelectCard
                key={runner.id ?? i}
                runner={runner}
                isSelected={runner.id === selectedRunnerId}
                onSelect={onSelectRunner}
              />
            ))}
          </div>
        )}
      </Card>

      {selectedRunnerId && (
        <>
          <Card>
            <Text variant="base" weight="strong">
              Recent jobs
            </Text>
            <Table
              columns={jobColumns}
              data={jobs}
              isLoading={jobsLoading}
              enableSearch={false}
              emptyMessage="No recent jobs found"
              skeletonRows={5}
            />
          </Card>

          <SandboxConfigPanel
            runnerId={selectedRunnerId}
            adminEmail={adminEmail}
            configs={configs}
            logTemplates={templates.log_templates ?? []}
            planTemplates={templates.plan_templates ?? []}
            onRefresh={onRefreshConfigs}
          />
        </>
      )}
    </AdminSection>
  )
}
