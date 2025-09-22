import { AppWorkflowsTable, NoActions, Notice, Pagination } from '@/components'
import { getActions } from '@/lib'

export const AppActions = async ({
  appId,
  orgId,
  limit = 10,
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
    data: actions,
    error,
    headers,
  } = await getActions({
    appId,
    limit,
    offset,
    orgId,
    q,
  })

  const pageData = {
    hasNext: headers?.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }
  return error ? (
    <Notice>Can&apos;t load actions: {error?.error}</Notice>
  ) : actions ? (
    <div className="flex flex-col gap-4 w-full">
      <AppWorkflowsTable appId={appId} orgId={orgId} workflows={actions} />
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
