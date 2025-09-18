import { InstallComponentDeploys, Pagination, Text } from '@/components'
import { getDeploysByComponentId } from '@/lib'
import type { TComponent } from '@/types'

export const Deploys = async ({
  component,
  installId,
  orgId,
  limit = 6,
  offset,
}: {
  component: TComponent
  installId: string
  orgId: string
  limit?: number
  offset?: string
}) => {
  const { data: deploys, headers } = await getDeploysByComponentId({
    componentId: component.id,
    installId,
    orgId,
    offset,
    limit,
  })

  const pageData = {
    hasNext: headers?.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }

  return deploys ? (
    <div className="flex flex-col gap-4 w-full">
      <InstallComponentDeploys
        component={component}
        initDeploys={deploys}
        shouldPoll
      />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={limit}
      />
    </div>
  ) : (
    <Text>Unable to load deploy history.</Text>
  )
}
