'use server'

import { InstallActionsTable as Table } from '@/components/actions/InstallActionsTable'
import { getInstallActionsLatestRuns } from '@/lib'

const LIMIT = 10

export async function InstallActionsTable({
  installId,
  orgId,
  offset,
  q,
  trigger_types,
}: {
  installId: string
  orgId: string
  offset: string
  q?: string
  trigger_types?: string
}) {
  const {
    data: actionsWithRuns,
    error,
    headers,
  } = await getInstallActionsLatestRuns({
    installId,
    limit: LIMIT,
    orgId,
    offset,
    q,
  })

  const pagination = {
    limit: Number(headers?.['x-nuon-page-limit'] ?? LIMIT),
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  if (error && !actionsWithRuns) {
    return (
      <div>
        <p>Could not load your actions.</p>
        <p>{error.error}</p>
      </div>
    )
  }

  return (
    <Table
      actionsWithRuns={actionsWithRuns}
      pagination={pagination}
      shouldPoll
    />
  )
}
