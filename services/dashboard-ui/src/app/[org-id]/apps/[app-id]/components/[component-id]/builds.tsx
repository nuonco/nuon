import { ComponentBuildHistory, Pagination } from '@/components'
import { getComponentBuilds } from '@/lib'

export const Builds = async ({
  componentId,
  orgId,
  limit = 6,
  offset,
}: {
  componentId: string
  orgId: string
  limit?: number
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
        componentId={componentId}
        initBuilds={builds || []}
        offset={pageData?.offset}
        limit={limit}
        shouldPoll
      />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={limit}
      />
    </div>
  )
}
