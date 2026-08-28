import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { InstallConfigVersionDetails } from '@/components/install-configs/InstallConfigVersionDetails'
import { panelTriggerClass } from '@/components/surfaces/panel-trigger'
import { getInstallConfigVersionDiff } from '@/lib'
import type { TInstallConfigVersion } from '@/types'
import { getDiffSummary } from '../config-diff-utils'

const LIMIT = 10

export interface IInstallConfigsTimeline {
  versions: TInstallConfigVersion[]
  isLoading?: boolean
  orgId?: string
  installId?: string
}

export const InstallConfigsTimeline = ({
  versions,
  isLoading,
  orgId,
  installId,
}: IInstallConfigsTimeline) => {
  if (isLoading) return <TimelineSkeleton eventCount={5} />

  if (!versions?.length) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No configs yet"
        emptyMessage="Config versions appear here after the first sync from git."
      />
    )
  }

  return (
    <Timeline<TInstallConfigVersion>
      events={versions}
      eventCount={LIMIT}
      getEventKey={(version) => version.id}
      pagination={{ hasNext: false, offset: 0, limit: LIMIT }}
      renderEvent={(version) => (
        <InstallConfigVersionEvent
          version={version}
          orgId={orgId}
          installId={installId}
        />
      )}
    />
  )
}

const InstallConfigVersionEvent = ({
  version,
  orgId,
  installId,
}: {
  version: TInstallConfigVersion
  orgId?: string
  installId?: string
}) => {
  const sync = version?.install_config_sync
  const commit = sync?.vcs_connection_commit

  const { data: diff, isLoading: isDiffLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'install-config-version-diff',
      orgId,
      installId,
      version?.id,
    ],
    queryFn: () =>
      getInstallConfigVersionDiff({
        orgId: orgId!,
        installId: installId!,
        versionId: version.id,
      }),
    enabled: !!orgId && !!installId && !!version?.id,
    retry: 1,
  })

  const diffSummary = diff?.key ? getDiffSummary(diff) : null

  return (
    <TimelineEvent
      createdAt={version?.created_at}
      status={version?.status?.status ?? 'unknown'}
      badge={{
        children: sync?.triggered_by ?? 'unknown',
        theme: 'neutral',
      }}
      title={
        <InstallConfigVersionDetails
          version={version}
          diff={diff}
          isDiffLoading={isDiffLoading}
          panelKey={`install-config-version-${version?.id}`}
          triggerButton={{
            variant: 'ghost',
            className: panelTriggerClass,
            children: commit?.message
              ? commit.message.split('\n')[0]?.trim()
              : 'Config sync',
          }}
        />
      }
      caption={<ID>{version?.id}</ID>}
      additionalCaption={
        <span className="flex items-center gap-2">
          <Badge size="sm" theme={version?.created ? 'success' : 'info'}>
            {version?.created ? 'Created' : 'Updated'}
          </Badge>
          {version?.file_path ? (
            <Text variant="subtext" theme="neutral" family="mono">
              {version.file_path}
            </Text>
          ) : null}
          {diffSummary ? (
            <ChangeCountSummary
              added={diffSummary.added}
              updated={diffSummary.changed}
              removed={diffSummary.removed}
            />
          ) : null}
        </span>
      }
    />
  )
}
