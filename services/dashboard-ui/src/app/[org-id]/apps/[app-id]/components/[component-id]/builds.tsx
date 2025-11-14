import { EmptyState } from '@/components/common/EmptyState'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { BuildTimeline } from '@/components/builds/BuildTimeline'
import { getComponentBuilds } from '@/lib'
import type { TComponent } from '@/types'

export const Builds = async ({
  component,
  limit = 10,
  offset,
  orgId,
}: {
  component: TComponent
  limit?: number
  offset: string
  orgId: string
}) => {
  const {
    data: builds,
    error,
    headers,
  } = await getComponentBuilds({
    componentId: component?.id,
    limit,
    offset,
    orgId,
  })

  const pagination = {
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  return error ? (
    <BuildsError />
  ) : builds?.length ? (
    <>
      <BuildTimeline
        initBuilds={builds}
        componentId={component?.id}
        componentName={component?.name}
        pagination={pagination}
        shouldPoll
      />
    </>
  ) : (
    <BuildsError
      title="No builds yet"
      message="Once youre component has builds they will appear here."
    />
  )
}

export const BuildsSkeleton = () => {
  return (
    <>
      <TimelineSkeleton eventCount={10} />
    </>
  )
}

export const BuildsError = ({
  message = 'We encountered an issue loading your component builds. Please try refreshing the page.',
  title = 'Unable to load builds',
}: {
  message?: string
  title?: string
}) => {
  return (
    <EmptyState variant="history" emptyMessage={message} emptyTitle={title} />
  )
}
