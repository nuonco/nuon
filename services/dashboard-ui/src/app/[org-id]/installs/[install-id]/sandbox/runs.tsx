import { EmptyState } from '@/components/common/EmptyState'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { SandboxRunsTimeline } from '@/components/sandbox/SandboxRunsTimeline'
import { getInstallSandboxRuns } from '@/lib'

export const Runs = async ({
  installId,
  limit = 10,
  offset,
  orgId,
}: {
  installId: string
  limit?: number
  offset: string
  orgId: string
}) => {
  const {
    data: runs,
    error,
    headers,
  } = await getInstallSandboxRuns({
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
  ) : runs?.length ? (
    <>
      <SandboxRunsTimeline initRuns={runs} pagination={pagination} shouldPoll />
    </>
  ) : (
    <RunsError
      title="No sandbox runs yet"
      message="Once the install is provisioned youre sandbox runs will appear here."
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
  message = 'We encountered an issue loading your sandbox runs. Please try refreshing the page.',
  title = 'Unable to load runs',
}: {
  message?: string
  title?: string
}) => {
  return (
    <EmptyState variant="history" emptyMessage={message} emptyTitle={title} />
  )
}
