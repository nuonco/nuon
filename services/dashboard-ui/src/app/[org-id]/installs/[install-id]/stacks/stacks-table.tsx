import { Banner } from '@/components/common/Banner'
import {
  InstallStacksTable as Table,
  InstallStacksTableSkeleton as Skeleton,
} from '@/components/stacks/InstallStacksTable'
import { getInstallStack } from '@/lib'

const LIMIT = 10

export const InstallStacksTable = async ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) => {
  const {
    data: stack,
    error,
    headers,
  } = await getInstallStack({
    installId,
    orgId,
  })

  const pagination = {
    limit: Number(headers?.['x-nuon-page-limit'] ?? LIMIT),
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  return error ? (
    <Banner theme="error">
      Can&apos;t load install stacks: {error?.error}
    </Banner>
  ) : (
    <Table stack={stack} pagination={pagination} shouldPoll />
  )
}

export const InstallStacksTableSkeleton = Skeleton
