'use client'

import type { ColumnDef } from '@tanstack/react-table'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { getAppBranches } from '@/lib'
import type { TAppBranch } from '@/types'
import { useEffect, useState } from 'react'
import { DataTable } from '@/components/old/DataTable'

interface IAppBranches {
  appId: string
  orgId: string
  offset?: string
}

type TBranchRow = {
  branchId: string
  branchName: string
  lastCommit: string | null
  workflowCount: number
  createdAt: string
  href: string
}

function parseBranchesToTableData(
  branches: TAppBranch[],
  orgId: string,
  appId: string
): TBranchRow[] {
  return branches.map((branch) => ({
    branchId: branch.id || '',
    branchName: branch.name || '',
    lastCommit: branch.last_synced_commit || null,
    workflowCount: branch.workflows?.length || 0,
    createdAt: branch.created_at || '',
    href: `/${orgId}/apps/${appId}/branches/${branch.id}`,
  }))
}

export const AppBranches = ({ appId, orgId, offset }: IAppBranches) => {
  const [branches, setBranches] = useState<TAppBranch[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchBranches = async () => {
      setLoading(true)
      const { data, error: fetchError } = await getAppBranches({
        appId,
        orgId,
        offset: offset ? Number(offset) : undefined,
      })

      if (fetchError) {
        setError(
          typeof fetchError === 'string'
            ? fetchError
            : fetchError.user_error ||
                fetchError.error ||
                fetchError.description ||
                'Failed to load branches'
        )
      } else {
        setBranches(data || [])
      }
      setLoading(false)
    }

    fetchBranches()
  }, [appId, orgId, offset])

  if (loading) {
    return <div>Loading branches...</div>
  }

  if (error) {
    return <div className="text-red-600">Can&apos;t load branches: {error}</div>
  }

  const headers = [
    { key: 'branchName', value: 'Branch Name' },
    { key: 'lastCommit', value: 'Last Synced Commit' },
    { key: 'workflowCount', value: 'Workflows' },
    { key: 'createdAt', value: 'Created' },
    { key: 'action', value: '' },
  ]

  const tableData = parseBranchesToTableData(branches, orgId, appId).map(
    (row) => ({
      branchName: (
        <span>
          <Link href={row.href}>{row.branchName}</Link>
          <div className="text-xs text-gray-500">{row.branchId}</div>
        </span>
      ),
      lastCommit: row.lastCommit ? (
        <code className="text-xs bg-gray-100 px-2 py-1 rounded">
          {row.lastCommit.slice(0, 7)}
        </code>
      ) : (
        <span className="text-gray-400">Not synced yet</span>
      ),
      workflowCount: `${row.workflowCount} workflow${row.workflowCount !== 1 ? 's' : ''}`,
      createdAt: <Time time={row.createdAt} format="relative" />,
      action: (
        <Link href={row.href}>
          View <Icon variant="CaretRightIcon" />
        </Link>
      ),
    })
  )

  return (
    <div className="w-full">
      <DataTable
        headers={headers}
        initData={tableData}
        emptyMessage="Create your first app branch to get started with version management."
      />
    </div>
  )
}