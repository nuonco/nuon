import { Pagination, Text } from '@/components'
import { InstallActionRuns } from '@/components/old/InstallActionRuns'
import { getInstallActionById } from '@/lib'

export const ActionRuns = async ({
  actionId,
  installId,
  orgId,
  limit = '6',
  offset,
}: {
  actionId: string
  installId: string
  orgId: string
  limit?: string
  offset?: string
}) => {
  const {
    data: installAction,
    error,
    headers,
  } = await getInstallActionById({
    orgId,
    installId,
    actionId,
    offset,
    limit,
  })

  const pageData = {
    hasNext: headers?.['x-nuon-page-next'] || 'false',
    offset: headers?.['x-nuon-page-offset'] || '0',
  }

  return installAction && !error ? (
    <div className="flex flex-col gap-4 w-full">
      <InstallActionRuns
        initInstallAction={installAction}
        pagination={{ offset: pageData.offset, limit }}
        shouldPoll
      />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={parseInt(limit)}
      />
    </div>
  ) : (
    <Text>Unable to load action run history.</Text>
  )
}
