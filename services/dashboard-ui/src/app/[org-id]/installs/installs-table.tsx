'use server'

import { InstallsTable as Table } from '@/components/installs/InstallsTable'
import { getInstalls } from '@/lib'

const LIMIT = 10

export async function InstallsTable({
  orgId,
  offset,
  q,
}: {
  orgId: string
  offset: string
  q?: string
}) {
  const {
    data: installs,
    error,
    headers,
  } = await getInstalls({ limit: LIMIT, orgId, offset, q })

  const pagination = {
    limit: Number(headers?.['x-nuon-page-limit'] ?? LIMIT),
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  if (error && !installs) {
    return (
      <div>
        <p>Could not load your installs.</p>
        <p>{error.error}</p>
      </div>
    )
  }

  return <Table installs={installs} pagination={pagination} shouldPoll />
}
