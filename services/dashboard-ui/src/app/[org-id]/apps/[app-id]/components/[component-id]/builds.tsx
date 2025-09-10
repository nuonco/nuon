import { ComponentBuildHistory, Pagination } from '@/components'
import { getComponentBuilds } from '@/lib'

export const Builds = async ({
  appId,
  componentId,
  orgId,
  limit = '6',
  offset,
}: {
  appId: string
  componentId: string
  orgId: string
  limit?: string
  offset?: string
}) => {
  const { data: builds, headers } = await getComponentBuilds({
    orgId,
    componentId,
    offset,
    limit,
  })

  const pageData = {
    hasNext: headers?.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }

  return (
    <div className="flex flex-col gap-4 w-full">
      <ComponentBuildHistory
        appId={appId}
        componentId={componentId}
        initBuilds={builds || []}
        orgId={orgId}
        shouldPoll
      />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={parseInt(limit)}
      />
    </div>
  )
}
