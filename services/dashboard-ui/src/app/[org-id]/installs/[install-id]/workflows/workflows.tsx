import { WorkflowTimeline } from '@/components/workflows/WorkflowTimeline'
import { Text } from '@/components/common/Text'
import { getInstallWorkflows } from '@/lib'

interface ILoadWorkflows {
  installId: string
  offset: string
  orgId: string
  showDrift?: boolean
  type?: string
}

export async function Workflows({
  orgId,
  offset,
  installId,
  showDrift = true,
  type = '',
}: ILoadWorkflows) {
  const {
    data: workflows,
    error,
    headers,
  } = await getInstallWorkflows({
    orgId,
    installId,
    offset,
    type,
    planonly: showDrift,
  })

  const pagination = {
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  return workflows && !error ? (
    <>
      <WorkflowTimeline
        initWorkflows={workflows}
        pagination={pagination}
        ownerId={installId}
        ownerType="installs"
        shouldPoll
        planonly={showDrift}
        type={type}
      />
    </>
  ) : (
    <WorkflowsError />
  )
}

export const WorkflowsError = () => (
  <div className="w-full">
    <Text>Error fetching recenty workflows activity </Text>
  </div>
)
