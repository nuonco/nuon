import {
  InstallActionWorkflowsTable,
  NoActions,
  Notice,
  Pagination,
} from '@/components'
import { getInstallActionsLatestRuns } from '@/lib'

export const InstallActions = async ({
  installId,
  orgId,
  limit = 10,
  offset,
  q,
  trigger_types,
}: {
  installId: string
  orgId: string
  limit?: number
  offset?: string
  q?: string
  trigger_types?: string
}) => {
  const {
    data: actionsWithLatestRun,
    error,
    headers,
  } = await getInstallActionsLatestRuns({
    installId,
    limit,
    offset,
    orgId,
    q,
    trigger_types,
  })

  const pageData = {
    hasNext: headers?.['x-nuon-page-next'] || 'false',
    offset: headers?.['x-nuon-page-offset'] || '0',
  }

  return error ? (
    <Notice className="grow-0 h-max">Can&apos;t load install actions</Notice>
  ) : actionsWithLatestRun ? (
    <div className="flex flex-col gap-4 w-full">
      <InstallActionWorkflowsTable
        actions={actionsWithLatestRun}
        installId={installId}
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
    <NoActions />
  )
}
