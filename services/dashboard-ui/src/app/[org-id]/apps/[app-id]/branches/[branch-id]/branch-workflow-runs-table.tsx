'use client'

import { useEffect, useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Link } from '@/components/common/Link'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Duration } from '@/components/common/Duration'
import { Table } from '@/components/common/Table'
import { Banner } from '@/components/common/Banner'
import type { TInstallWorkflow } from '@/types'
import { getBranchWorkflowRuns } from '@/lib'

type TWorkflowRunRow = {
  runId: string
  workflowId: string
  status: string
  configVersion: number | null
  startedAt: string
  completedAt: string | null
  duration: number | null
  errorMessage: string | null
  href: string
}

function parseWorkflowRunsToTableData(
  runs: TInstallWorkflow[],
  orgId: string,
  appId: string,
  branchId: string
): TWorkflowRunRow[] {
  return runs.map((workflow) => {
    // Get the first app branch run (there should only be one per workflow)
    const branchRun = workflow.app_branch_runs?.[0]
    
    // If no branch run data exists, use workflow data as fallback
    if (!branchRun) {
      console.warn('Workflow missing app_branch_runs data:', workflow.id)
      return {
        runId: workflow.id || '',
        workflowId: workflow.id || '',
        status: workflow.status || 'unknown',
        configVersion: null,
        startedAt: workflow.created_at || '',
        completedAt: null,
        duration: workflow.created_at && workflow.updated_at 
          ? new Date(workflow.updated_at).getTime() - new Date(workflow.created_at).getTime()
          : null,
        errorMessage: null,
        href: `/${orgId}/apps/${appId}/branches/${branchId}/runs/${workflow.id}`,
      }
    }
    
    // Use branch run data if available, otherwise fall back to workflow data
    const status = branchRun.status || workflow.status || 'unknown'
    const startedAt = branchRun.started_at || workflow.created_at || ''
    const completedAt = branchRun.completed_at || null
    
    // Calculate duration from branch run or workflow timestamps
    let duration = null
    if (branchRun.started_at && branchRun.completed_at) {
      duration = new Date(branchRun.completed_at).getTime() - new Date(branchRun.started_at).getTime()
    } else if (workflow.created_at && workflow.updated_at) {
      duration = new Date(workflow.updated_at).getTime() - new Date(workflow.created_at).getTime()
    }

    return {
      runId: branchRun.id || workflow.id || '',
      workflowId: workflow.id || '',
      status,
      configVersion: branchRun.app_branch_config?.config_number || null,
      startedAt,
      completedAt,
      duration,
      errorMessage: branchRun.error_message || null,
      href: `/${orgId}/apps/${appId}/branches/${branchId}/runs/${workflow.id}`,
    }
  })
}

const columns: ColumnDef<TWorkflowRunRow>[] = [
  {
    accessorKey: 'runId',
    header: 'Run ID',
    cell: (info) => (
      <Link href={info.row.original.href}>
        <code className="text-xs">{(info.getValue() as string).slice(0, 8)}...</code>
      </Link>
    ),
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: (info) => {
      const status = info.getValue() as string
      const errorMessage = info.row.original.errorMessage
      return (
        <div className="flex flex-col gap-1">
          <Status status={status} />
          {errorMessage && status === 'failed' && (
            <Text variant="subtext" theme="error" className="text-xs">
              {errorMessage.slice(0, 50)}{errorMessage.length > 50 ? '...' : ''}
            </Text>
          )}
        </div>
      )
    },
    enableSorting: true,
  },
  {
    accessorKey: 'configVersion',
    header: 'Config',
    cell: (info) => {
      const version = info.getValue() as number | null
      return version ? (
        <Text variant="body">v{version}</Text>
      ) : (
        <Text variant="subtext" theme="neutral">
          -
        </Text>
      )
    },
  },
  {
    accessorKey: 'startedAt',
    header: 'Started',
    cell: (info) => <Time time={info.getValue() as string} format="relative" />,
    enableSorting: true,
  },
  {
    accessorKey: 'duration',
    header: 'Duration',
    cell: (info) => {
      const duration = info.getValue() as number | null
      return duration ? (
        <Duration nanoseconds={duration * 1e6} />
      ) : (
        <Text variant="subtext" theme="neutral">
          --
        </Text>
      )
    },
  },
  {
    enableSorting: false,
    accessorKey: 'href',
    id: 'action',
    header: '',
    cell: (info) => (
      <Text>
        <Link className="text-left" href={info.getValue() as string}>
          View <Icon variant="CaretRightIcon" />
        </Link>
      </Text>
    ),
  },
]

interface IBranchWorkflowRunsTable {
  appId: string
  branchId: string
  orgId: string
}

export const BranchWorkflowRunsTable = ({
  appId,
  branchId,
  orgId,
}: IBranchWorkflowRunsTable) => {
  const [runs, setRuns] = useState<TInstallWorkflow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchRuns = async () => {
      setLoading(true)
      const { data, error: fetchError } = await getBranchWorkflowRuns({
        appId,
        branchId,
        orgId,
      })

      if (fetchError) {
        console.error('Failed to fetch workflow runs:', fetchError)
        setError(
          typeof fetchError === 'string'
            ? fetchError
            : fetchError.user_error ||
                fetchError.error ||
                fetchError.description ||
                'Failed to load workflow runs'
        )
      } else {
        // Sort by created_at descending (newest first)
        const sorted = (data || []).sort(
          (a, b) =>
            new Date(b.created_at || '').getTime() -
            new Date(a.created_at || '').getTime()
        )
        setRuns(sorted)
      }
      setLoading(false)
    }

    fetchRuns()
  }, [appId, branchId, orgId])

  if (loading) {
    return (
      <div className="rounded-lg border border-gray-200 dark:border-gray-700">
        <div className="h-12 bg-gray-100 dark:bg-gray-800 rounded-t-lg animate-pulse" />
        {Array.from({ length: 5 }).map((_, i) => (
          <div
            key={i}
            className="h-16 bg-gray-50 dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 animate-pulse"
          />
        ))}
      </div>
    )
  }

  if (error) {
    return <Banner theme="error">Can&apos;t load workflow runs: {error}</Banner>
  }

  return (
    <Table<TWorkflowRunRow>
      columns={columns}
      data={parseWorkflowRunsToTableData(runs, orgId, appId, branchId)}
      emptyStateProps={{
        emptyMessage: 'No workflow runs yet. Trigger a run to see activity here.',
        emptyTitle: 'No workflow runs',
      }}
      searchPlaceholder="Search workflow runs..."
    />
  )
}