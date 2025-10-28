import { Banner } from '@/components/common/Banner'
import {
  AppInstallsTable as Table,
  AppInstallsTableSkeleton as Skeleton,
} from '@/components/apps/AppInstallsTable'
import { getInstallsByAppId } from '@/lib'

const LIMIT = 10

export const InstallsTable = async ({
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
    data: installs,
    error,
    headers,
  } = await getInstallsByAppId({
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
    <Banner theme="error">Can&apos;t load installs: {error?.error}</Banner>
  ) : (
    <Table appId={appId} installs={installs} pagination={pagination} shouldPoll />
  )
}

export const InstallsTableSkeleton = Skeleton

