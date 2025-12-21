import { Banner } from '@/components/common/Banner'
import { getAppBranches } from '@/lib'
import { BranchesTableClient } from './branches-table-client'

const LIMIT = 10

export const BranchesTable = async ({
  appId,
  orgId,
  limit = LIMIT,
  offset,
}: {
  appId: string
  orgId: string
  limit?: number
  offset?: string
}) => {
  const {
    data: branches,
    error,
    headers,
  } = await getAppBranches({
    appId,
    limit,
    offset: offset ? Number(offset) : undefined,
    orgId,
  })

  const pagination = {
    limit: Number(headers?.['x-nuon-page-limit'] ?? LIMIT),
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  return error ? (
    <Banner theme="error">
      Can&apos;t load branches:{' '}
      {typeof error === 'string'
        ? error
        : error?.error || error?.description || 'Unknown error'}
    </Banner>
  ) : (
    <BranchesTableClient branches={branches || []} pagination={pagination} />
  )
}

export const BranchesTableSkeleton = () => {
  return (
    <div className="flex flex-col gap-4 w-full">
      <div className="h-10 bg-gray-200 dark:bg-gray-700 animate-pulse rounded" />
      <div className="rounded-lg border border-gray-200 dark:border-gray-700">
        <div className="h-12 bg-gray-100 dark:bg-gray-800 rounded-t-lg" />
        {Array.from({ length: 5 }).map((_, i) => (
          <div
            key={i}
            className="h-16 bg-gray-50 dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 animate-pulse"
          />
        ))}
      </div>
    </div>
  )
}
