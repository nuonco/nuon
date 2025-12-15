import { EmptyState } from '@/components/common/EmptyState'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { DeployTimeline } from '@/components/deploys/DeployTimeline'
import { getComponentDeploys } from '@/lib'
import type { TComponent } from '@/types'

export const Deploys = async ({
  component,
  installId,
  limit = 10,
  offset,
  orgId,
}: {
  component: TComponent
  installId: string
  limit?: number
  offset: string
  orgId: string
}) => {
  const {
    data: deploys,
    error,
    headers,
  } = await getComponentDeploys({
    componentId: component?.id,
    installId,
    limit,
    offset,
    orgId,
  })

  const pagination = {
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  return error ? (
    <DeploysError />
  ) : deploys?.length ? (
    <>
      <DeployTimeline
        initDeploys={deploys}
        componentId={component?.id}
        componentName={component?.name}
        pagination={pagination}
        shouldPoll
      />
    </>
  ) : (
    <DeploysError
      title="No deploys yet"
      message="Once the install is provisioned youre component deploys will appear here."
    />
  )
}

export const DeploysSkeleton = () => {
  return (
    <>
      <TimelineSkeleton eventCount={10} />
    </>
  )
}

export const DeploysError = ({
  message = 'We encountered an issue loading your component deploys. Please try refreshing the page.',
  title = 'Unable to load deploys',
}: {
  message?: string
  title?: string
}) => {
  return (
    <EmptyState variant="history" emptyMessage={message} emptyTitle={title} />
  )
}
