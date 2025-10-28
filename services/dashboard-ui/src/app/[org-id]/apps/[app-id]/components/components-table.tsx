import { Banner } from '@/components/common/Banner'
import {
  ComponentsTable as Table,
  ComponentsTableSkeleton as Skeleton,
} from '@/components/components/ComponentsTable'
import { getComponents } from '@/lib'

const LIMIT = 10

export const ComponentsTable = async ({
  appId,
  orgId,
  limit = LIMIT,
  offset,
  q,
  types,
}: {
  appId: string
  orgId: string
  limit?: number
  offset?: string
  q?: string
  types?: string
}) => {
  const {
    data: components,
    error,
    headers,
  } = await getComponents({
    appId,
    limit,
    offset,
    orgId,
    q,
    types,
  })

  const pagination = {
    limit: Number(headers?.['x-nuon-page-limit'] ?? LIMIT),
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  return error ? (
    <Banner theme="error">Can&apos;t load components: {error?.error}</Banner>
  ) : (
    <Table components={components} pagination={pagination} shouldPoll />
  )
}

export const ComponentsTableSkeleton = Skeleton
