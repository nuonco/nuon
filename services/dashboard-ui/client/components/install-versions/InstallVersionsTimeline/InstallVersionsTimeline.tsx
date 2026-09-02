import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import {
  InstallVersionDetails,
  installVersionSource,
  resolveInstallVersionStatus,
} from '@/components/install-versions/InstallVersionDetails'
import { panelTriggerClass } from '@/components/surfaces/panel-trigger'
import type { TInstallAppConfigVersion } from '@/types'

const LIMIT = 10

export interface IInstallVersionsTimeline {
  versions: TInstallAppConfigVersion[]
  isLoading?: boolean
  orgId?: string
  installId?: string
  appId?: string
}

export const InstallVersionsTimeline = ({
  versions,
  isLoading,
  orgId,
  installId,
  appId,
}: IInstallVersionsTimeline) => {
  if (isLoading) return <TimelineSkeleton eventCount={5} />

  if (!versions?.length) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No versions yet"
        emptyMessage="App config versions appear here after the first sync."
      />
    )
  }

  return (
    <Timeline<TInstallAppConfigVersion>
      events={versions}
      eventCount={LIMIT}
      getEventKey={(version) => version.id}
      pagination={{ hasNext: false, offset: 0, limit: LIMIT }}
      renderEvent={(version) => {
        const branchRun = version?.app_branch_run
        const commit = branchRun?.vcs_connection_commit

        return (
          <TimelineEvent
            createdAt={version?.created_at}
            status={resolveInstallVersionStatus(version)}
            badge={{
              children: installVersionSource(version),
              theme: 'neutral',
            }}
            title={
              <InstallVersionDetails
                version={version}
                orgId={orgId}
                installId={installId}
                appId={appId}
                panelKey={`install-version-${version?.id}`}
                triggerButton={{
                  variant: 'ghost',
                  className: panelTriggerClass,
                  children: commit?.message
                    ? commit.message.split('\n')[0]?.trim()
                    : 'App config update',
                }}
              />
            }
            caption={<ID>{version?.new_app_config_id || version?.id}</ID>}
            additionalCaption={
              commit ? (
                <span className="flex items-center gap-2">
                  {branchRun?.app_branch?.name ? (
                    <Badge size="sm" theme="info">
                      {branchRun.app_branch.name}
                    </Badge>
                  ) : null}
                  {commit.sha ? (
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
  )
}
