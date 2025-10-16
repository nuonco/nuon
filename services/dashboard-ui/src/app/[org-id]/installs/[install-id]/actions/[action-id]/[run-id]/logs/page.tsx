import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { getInstallActionById, getInstallActionRunById } from '@/lib'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import type { TPageProps } from '@/types'
import { Logs, LogsError, LogsSkeleton } from './logs'

type TInstallPageProps = TPageProps<
  'org-id' | 'install-id' | 'action-id' | 'run-id'
>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionId,
    ['run-id']: runId,
  } = await params
  const [{ data: installActionRun }, { data: installAction }] =
    await Promise.all([
      getInstallActionRunById({
        installId,
        orgId,
        runId,
      }),
      getInstallActionById({
        actionId,
        installId,
        orgId,
      }),
    ])

  return {
    title: `Logs | ${installAction?.action_workflow?.name} | ${installActionRun.trigger_type} run`,
  }
}

export default async function InstallAcitonRunLogsPage({
  params,
}: TInstallPageProps) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionId,
    ['run-id']: runId,
  } = await params
  const [{ data: installActionRun }] = await Promise.all([
    getInstallActionRunById({
      installId,
      orgId,
      runId,
    }),
  ])

  return (
    <>
      <LogStreamProvider
        initLogStream={installActionRun?.log_stream}
        shouldPoll={installActionRun?.log_stream?.open}
      >
        <ErrorBoundary fallback={<LogsError />}>
          <Suspense fallback={<LogsSkeleton />}>
            <Logs
              logStreamId={installActionRun?.log_stream?.id}
              orgId={orgId}
            />
          </Suspense>
        </ErrorBoundary>
      </LogStreamProvider>
    </>
  )
}
