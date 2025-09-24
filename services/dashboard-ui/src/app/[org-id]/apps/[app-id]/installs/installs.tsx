import { AppInstallsTable, Notice, NoInstalls, Pagination } from '@/components'
import { getInstallsByAppId } from '@/lib'
import type { TApp } from '@/types'

export const AppInstalls = async ({
  app,
  appId,
  orgId,
  limit = 10,
  offset,
  q,
}: {
  app: TApp
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

  const pageData = {
    hasNext: headers?.['x-nuon-page-next'] || 'false',
    offset: headers?.['x-nuon-page-offset'] || '0',
  }

  return error ? (
    <Notice>Can&apos;t load installs: {error?.error}</Notice>
  ) : installs ? (
    <div className="flex flex-col gap-8 w-full">
      <AppInstallsTable
        installs={installs.map((install) => ({ ...install, app }))}
        orgId={orgId}
      />
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
