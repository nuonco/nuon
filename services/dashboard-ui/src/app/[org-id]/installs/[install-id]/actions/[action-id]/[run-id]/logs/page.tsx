import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import {
  getInstallAction,
  getInstallActionRun,
  getInstall,
  getOrg,
} from '@/lib'
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
      getInstallActionRun({
        installId,
        orgId,
        runId,
      }),
      getInstallAction({
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
  const [
    { data: installActionRun },
    { data: installAction },
    { data: install },
    { data: org },
  ] = await Promise.all([
    getInstallActionRun({
      installId,
      orgId,
      runId,
    }),
    getInstallAction({
      actionId,
      installId,
      orgId,
    }),
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
  ])

  return (
    <>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name,
          },
          {
            path: `/${orgId}/installs/${installId}/actions`,
            text: 'Actions',
          },
          {
            path: `/${orgId}/installs/${installId}/actions/${actionId}`,
            text: installAction?.action_workflow?.name,
          },
          {
            path: `/${orgId}/installs/${installId}/actions/${actionId}/${runId}`,
            text: `${installActionRun?.trigger_type} run`,
          },
          {
            path: `/${orgId}/installs/${installId}/actions/${actionId}/${runId}/logs`,
            text: `Logs`,
          },
        ]}
      />
      <LogStreamProvider
        initLogStream={installActionRun?.log_stream}
        shouldPoll={installActionRun?.log_stream?.open}
      >
        <ErrorBoundary fallback={<LogsError />}>
          <Suspense fallback={<LogsSkeleton />}>
            <Logs
              actionConfig={installActionRun?.config}
              logStreamId={installActionRun?.log_stream?.id}
              logStreamOpen={installActionRun?.log_stream?.open}
              orgId={orgId}
            />
          </Suspense>
        </ErrorBoundary>
      </LogStreamProvider>
    </>
  )
}
