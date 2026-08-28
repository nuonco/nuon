import { useMutation, useQueryClient } from '@tanstack/react-query'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import { approveInstallCreation } from '@/lib'
import type { TAPIError, TAppInstallConfigSync } from '@/types'

const LIMIT = 10

export interface IAppInstallSyncsTimeline {
  syncs: TAppInstallConfigSync[]
  isLoading?: boolean
  orgId?: string
  appId?: string
}

export const AppInstallSyncsTimeline = ({
  syncs,
  isLoading,
  orgId,
  appId,
}: IAppInstallSyncsTimeline) => {
  if (isLoading) return <TimelineSkeleton eventCount={5} />

  if (!syncs?.length) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No install syncs yet"
        emptyMessage="Configure an installs config source and trigger a sync to see runs here."
      />
    )
  }

  return (
    <Timeline<TAppInstallConfigSync>
      events={syncs}
      eventCount={LIMIT}
      getEventKey={(sync) => sync.id}
      pagination={{ hasNext: false, offset: 0, limit: LIMIT }}
      renderEvent={(sync) => (
        <AppInstallSyncEvent sync={sync} orgId={orgId} appId={appId} />
      )}
    />
  )
}

const AppInstallSyncEvent = ({
  sync,
  orgId,
  appId,
}: {
  sync: TAppInstallConfigSync
  orgId?: string
  appId?: string
}) => {
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const status = sync?.status?.status || 'unknown'
  const approval = sync?.install_creation_approval
  const isAwaitingApproval =
    status === 'awaiting_approval' && approval?.status === 'pending'
  const proposedCount = approval?.proposed_installs?.length
  const installCount = sync?.install_config_syncs?.length

  const { mutate: approve, isPending: isApproving } = useMutation({
    mutationFn: (approvalId: string) =>
      approveInstallCreation({
        appId: appId!,
        syncId: sync.id,
        approvalId,
        orgId: orgId!,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['app-install-syncs', orgId, appId],
      })
      addToast(
        <Toast heading="Installs approved" theme="success">
          <Text>Creating missing installs and re-running sync.</Text>
        </Toast>
      )
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Approval failed" theme="error">
          <Text>{err?.error || 'Unable to approve install creation.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <TimelineEvent
      createdAt={sync?.created_at}
      status={status}
      title={
        orgId && appId ? (
          <Link
            href={`/${orgId}/apps/${appId}/install-syncs/${sync?.id}`}
            variant="inline"
          >
            Sync triggered by {sync?.triggered_by}
          </Link>
        ) : (
          `Sync triggered by ${sync?.triggered_by}`
        )
      }
      caption={
        sync?.vcs_connection_commit?.sha ? (
          <Text variant="subtext" theme="neutral" family="mono">
            {sync.vcs_connection_commit.sha.slice(0, 7)}
          </Text>
        ) : null
      }
      additionalCaption={
        <span className="flex flex-wrap items-center gap-2">
          {isAwaitingApproval ? (
            <Badge size="sm" theme="warn">
              {typeof proposedCount === 'number'
                ? `${proposedCount} new install${proposedCount === 1 ? '' : 's'} need approval`
                : 'Needs approval'}
            </Badge>
          ) : null}
          {installCount ? (
            <Text variant="subtext" theme="neutral">
              {installCount} install{installCount === 1 ? '' : 's'}
            </Text>
          ) : null}
          {sync?.status?.status_human_description ? (
            <Text variant="subtext" theme="neutral">
              {sync.status.status_human_description}
            </Text>
          ) : null}
          {sync?.queue_id && sync?.queue_signal_id ? (
            <AdminDashboardLink
              path={`/queues/${sync.queue_id}/signals/${sync.queue_signal_id}`}
              label="View signal"
            />
          ) : null}
        </span>
      }
      actions={
        isAwaitingApproval && approval?.id ? (
          <Button
            variant="primary"
            disabled={isApproving}
            onClick={() => approve(approval.id!)}
          >
            {isApproving ? 'Approving installs' : 'Approve creation'}
          </Button>
        ) : null
      }
    />
  )
}
