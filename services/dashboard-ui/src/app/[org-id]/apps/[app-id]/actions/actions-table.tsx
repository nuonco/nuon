import { Banner } from '@/components/common/Banner'
import {
  ActionsTable as Table,
  ActionsTableSkeleton as Skeleton,
} from '@/components/actions/ActionsTable'
import { getActions } from '@/lib'

const LIMIT = 10

export const ActionsTable = async ({
  appId,
  orgId,
  limit = LIMIT,
  offset,
  q,
}: {
  appId: string
  orgId: string
  limit?: number
  offset?: string
  q?: string
}) => {
  const {
    data: actions,
    error,
    headers,
  } = await getActions({
    appId,
    limit,
    offset,
    orgId,
    q,
  })

  const pagination = {
    limit: Number(headers?.['x-nuon-page-limit'] ?? LIMIT),
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  return error ? (
    <Banner theme="error">Can&apos;t load actions: {error?.error}</Banner>
  ) : (
    <Table actions={actions} pagination={pagination} shouldPoll />
  )
}

export const ActionsTableSkeleton = Skeleton
