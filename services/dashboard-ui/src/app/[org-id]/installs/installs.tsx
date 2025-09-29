import { NoInstalls, OrgInstallsTable, Notice, Pagination } from '@/components'
import { getInstalls } from '@/lib'

export const Installs = async ({
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
    data: installs,
    error,
    headers,
  } = await getInstalls({
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
    <Notice>Can&apos;t load installs: {error?.error}</Notice>
  ) : installs ? (
    <div className="flex flex-col gap-4 w-full">
      <OrgInstallsTable orgId={orgId} installs={installs} />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={limit}
      />
    </div>
  ) : (
    <NoInstalls />
  )
}
