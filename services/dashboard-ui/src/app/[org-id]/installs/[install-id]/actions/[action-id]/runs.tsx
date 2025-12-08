import { EmptyState } from '@/components/common/EmptyState'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { InstallActionRunTimeline } from '@/components/actions/InstallActionRunTimeline'
import { getInstallActionById } from '@/lib'

export const Runs = async ({
  actionId,
  installId,
  limit = 10,
  offset,
  orgId,
}: {
  actionId: string
  installId: string
  limit?: number
  offset: string
  orgId: string
}) => {
  const {
    data: action,
    error,
    headers,
  } = await getInstallActionById({
    actionId,
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
    <RunsError />
  ) : action ? (
    <>
      <InstallActionRunTimeline
        initInstallAction={action}
        pagination={pagination}
        shouldPoll
      />
    </>
  ) : (
    <RunsError
      title="No action runs yet"
      message="Once this action has been executed the run history will appear here."
    />
  )
}

export const RunsSkeleton = () => {
  return (
    <>
      <TimelineSkeleton eventCount={10} />
    </>
  )
}

export const RunsError = ({
  message = 'We encountered an issue loading your action runs. Please try refreshing the page.',
  title = 'Unable to load action runs',
}: {
  message?: string
  title?: string
}) => {
  return (
    <EmptyState variant="history" emptyMessage={message} emptyTitle={title} />
  )
}
