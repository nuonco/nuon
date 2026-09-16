import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { panelTriggerClass } from '@/components/surfaces/panel-trigger'
import type { TInstallUpdate } from '@/types'
import { humanize } from '@/utils/string-utils'
import {
  InstallUpdateDetails,
  installUpdateStatus,
} from '@/components/install-updates/InstallUpdateDetails'

const PAGE_SIZE = 20

export interface IInstallUpdatesTimeline {
  updates: TInstallUpdate[]
  isLoading?: boolean
  hasMore?: boolean
  onLoadMore?: () => void
  orgId?: string
  installId?: string
  appId?: string
}

const updateTitle = (update: TInstallUpdate): string => {
  const version = update.app_config?.version
  const branchRun = version?.app_branch_run
  const commit = branchRun?.vcs_connection_commit
  if (commit?.message)
    return commit.message.split('\n')[0]?.trim() || 'App config update'
  switch (update.type) {
    case 'inputs':
      return 'Inputs updated'
    case 'stack':
      return 'Stack version generated'
    case 'install_config':
      return 'Install config updated'
    default:
      return 'App config updated'
  }
}

export const InstallUpdatesTimeline = ({
  updates,
  isLoading,
  hasMore,
  onLoadMore,
  orgId,
  installId,
  appId,
}: IInstallUpdatesTimeline) => {
  if (isLoading) return <TimelineSkeleton eventCount={5} />

  if (!updates?.length) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No updates yet"
        emptyMessage="Updates appear here after app config versions are applied to this install."
      />
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <Timeline<TInstallUpdate>
        events={updates}
        eventCount={PAGE_SIZE}
        getEventKey={(update) => update.id}
        pagination={{ hasNext: false, offset: 0, limit: PAGE_SIZE }}
        renderEvent={(update) => {
          const version = update.app_config?.version
          const branchRun = version?.app_branch_run
          const commit = branchRun?.vcs_connection_commit

          return (
            <TimelineEvent
              createdAt={update.created_at}
              status={installUpdateStatus(update)}
              badge={{ children: humanize(update.type), theme: 'neutral' }}
              title={
                <InstallUpdateDetails
                  update={update}
                  orgId={orgId}
                  installId={installId}
                  appId={appId}
                  panelKey={`install-update-${update.id}`}
                  triggerButton={{
                    variant: 'ghost',
                    className: panelTriggerClass,
                    children: updateTitle(update),
                  }}
                />
              }
              caption={<ID>{update.id}</ID>}
              additionalCaption={
                commit || branchRun?.app_branch?.name ? (
                  <span className="flex items-center gap-2">
                    {branchRun?.app_branch?.name ? (
                      <Badge size="sm" theme="info">
                        {branchRun.app_branch.name}
                      </Badge>
                    ) : null}
                    {commit?.sha ? (
                      <Text variant="subtext" family="mono" theme="neutral">
                        {commit.sha.slice(0, 7)}
                      </Text>
                    ) : null}
                    {branchRun?.pr_number ? (
                      <Badge size="sm" theme="neutral">
                        PR #{branchRun.pr_number}
                      </Badge>
                    ) : null}
                  </span>
                ) : null
              }
            />
          )
        }}
      />
      {hasMore && onLoadMore ? (
        <Button className="self-center" onClick={onLoadMore}>
          Load more
        </Button>
      ) : null}
    </div>
  )
}
