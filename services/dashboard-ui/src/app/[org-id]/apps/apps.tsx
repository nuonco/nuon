import { NoApps, Notice, OrgAppsTable, Pagination } from '@/components'
import { getApps } from '@/lib'

export const Apps = async ({
  orgId,
  limit = 10,
  offset,
  q,
}: {
  orgId: string
  limit?: number
  offset?: string
  q?: string
}) => {
  const {
    data: apps,
    error,
    headers,
  } = await getApps({
    orgId,
    offset,
    limit,
    q,
  })

  const pageData = {
    hasNext: headers?.['x-nuon-page-next'] || 'false',
    offset: headers?.['x-nuon-page-offset'] || '0',
  }

  return error ? (
    <Notice>Can&apos;t load apps: {error?.error}</Notice>
  ) : apps ? (
    <div className="flex flex-col gap-4 w-full">
      <OrgAppsTable apps={apps} orgId={orgId} />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={limit}
      />
    </div>
  ) : (
    <NoApps />
  )
}
