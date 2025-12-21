'use client'

import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Table } from '@/components/common/Table'
import type { TAppBranchConfig } from '@/types'
import { getBranchConfigs } from '@/lib'
import { useEffect, useState } from 'react'
import { Banner } from '@/components/common/Banner'

type TConfigRow = {
  version: number
  repo: string | null
  branch: string | null
  groupCount: number
  createdAt: string
  isCurrent: boolean
}

function parseConfigsToTableData(
  configs: TAppBranchConfig[],
  currentConfigId?: string
): TConfigRow[] {
  return configs.map((config) => ({
    version: config.config_number || 0,
    repo: config.connected_github_vcs_config?.repo || null,
    branch: config.connected_github_vcs_config?.branch || null,
    groupCount: config.install_groups?.length || 0,
    createdAt: config.created_at || '',
    isCurrent: config.id === currentConfigId,
  }))
}

const columns: ColumnDef<TConfigRow>[] = [
  {
    accessorKey: 'version',
    header: 'Version',
    cell: (info) => (
      <div className="flex items-center gap-2">
        <Text variant="body" weight="strong">
          v{info.getValue() as number}
        </Text>
        {info.row.original.isCurrent && (
          <Badge theme="success" size="sm">
            CURRENT
          </Badge>
        )}
      </div>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'repo',
    header: 'Repository',
    cell: (info) => {
      const repo = info.getValue() as string | null
      return repo ? (
        <Text variant="body">{repo}</Text>
      ) : (
        <Text variant="subtext" theme="neutral">
          -
        </Text>
      )
    },
  },
  {
    accessorKey: 'branch',
    header: 'Branch',
    cell: (info) => {
      const branch = info.getValue() as string | null
      return branch ? (
        <code className="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
          {branch}
        </code>
      ) : (
        <Text variant="subtext" theme="neutral">
          -
        </Text>
      )
    },
  },
  {
    accessorKey: 'groupCount',
    header: 'Install Groups',
    cell: (info) => (
      <Text variant="body">
        {info.getValue() as number} group{(info.getValue() as number) !== 1 ? 's' : ''}
      </Text>
    ),
  },
  {
    accessorKey: 'createdAt',
    header: 'Created',
    cell: (info) => <Time time={info.getValue() as string} format="relative" />,
    enableSorting: true,
  },
]

interface IBranchConfigHistoryTable {
  appId: string
  branchId: string
  orgId: string
  currentConfigId?: string
}

export const BranchConfigHistoryTable = ({
  appId,
  branchId,
  orgId,
  currentConfigId,
}: IBranchConfigHistoryTable) => {
  const [configs, setConfigs] = useState<TAppBranchConfig[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchConfigs = async () => {
      setLoading(true)
      const { data, error: fetchError } = await getBranchConfigs({
        appId,
        branchId,
        orgId,
      })

      if (fetchError) {
        setError(
          typeof fetchError === 'string'
            ? fetchError
            : fetchError.user_error ||
                fetchError.error ||
                fetchError.description ||
                'Failed to load configurations'
        )
      } else {
        // Sort by config_number descending (newest first)
        const sorted = (data || []).sort(
          (a, b) => (b.config_number || 0) - (a.config_number || 0)
        )
        setConfigs(sorted)
      }
      setLoading(false)
    }

    fetchConfigs()
  }, [appId, branchId, orgId])

  if (loading) {
    return (
      <div className="rounded-lg border border-gray-200 dark:border-gray-700">
        <div className="h-12 bg-gray-100 dark:bg-gray-800 rounded-t-lg animate-pulse" />
        {Array.from({ length: 3 }).map((_, i) => (
          <div
            key={i}
            className="h-16 bg-gray-50 dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 animate-pulse"
          />
        ))}
      </div>
    )
  }

  if (error) {
    return <Banner theme="error">Can&apos;t load configuration history: {error}</Banner>
  }

  return (
    <Table<TConfigRow>
      columns={columns}
      data={parseConfigsToTableData(configs, currentConfigId)}
      emptyStateProps={{
        emptyMessage: 'No configuration versions yet.',
        emptyTitle: 'No configurations',
      }}
      searchPlaceholder="Search configurations..."
    />
  )
}