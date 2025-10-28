import { AppsTable as Table } from '@/components/apps/AppsTable'
import { Link } from '@/components/common/Link'
import { getApps } from '@/lib'

const LIMIT = 10

export async function AppsTable({
  orgId,
  offset,
  q,
}: {
  orgId: string
  offset: string
  q?: string
}) {
  const {
    data: apps,
    error,
    headers,
  } = await getApps({ limit: LIMIT, orgId, offset, q })

  const pagination = {
    limit: Number(headers?.['x-nuon-page-limit'] ?? LIMIT),
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  if (error && !apps) {
    return (
      <div>
        <p>Could not load your apps.</p>
        <p>{error.error}</p>
        <Link href="/api/auth/logout">Log out</Link>
      </div>
    )
  }

  return <Table apps={apps} pagination={pagination} shouldPoll />
}
