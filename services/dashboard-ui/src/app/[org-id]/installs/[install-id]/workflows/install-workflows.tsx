import { InstallWorkflowHistory, Pagination, Text } from '@/components'
import { getInstallWorkflows } from '@/lib'

export const InstallWorkflows = async ({
  installId,
  orgId,
  offset,
  limit = 20,
}: {
  installId: string
  orgId: string
  offset?: string
  limit?: number
}) => {
  const {
    data: workflows,
    error,
    headers,
  } = await getInstallWorkflows({ orgId, installId, offset, limit })

  const pageData = {
    hasNext: headers?.['x-nuon-page-next'] || 'false',
    offset: headers?.['x-nuon-page-offset'] || '0',
  }

  return workflows && !error ? (
    <div className="flex flex-col gap-4">
      <InstallWorkflowHistory
        initWorkflows={workflows}
        shouldPoll
        pagination={{ offset: pageData.offset, limit }}
      />
      <Pagination
        param="workflows"
        pageData={pageData}
        position="center"
        limit={limit}
      />
    </div>
  ) : (
    <Text>No install history yet.</Text>
  )
}
