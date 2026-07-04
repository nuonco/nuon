import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import {
  getInstallAppConfigVersions,
  getInstallAppConfigVersionDiff,
} from '@/lib'
import type { TInstallConfigDiff } from '@/lib/ctl-api/installs/get-install-app-config-version-diff'
import { getStatusTheme } from '@/utils/status-utils'
import type { TInstallAppConfigVersion } from '@/types'

export const Versions = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: versions, isLoading } = useQuery({
    queryKey: ['install-app-config-versions', org?.id, install?.id],
    queryFn: () =>
      getInstallAppConfigVersions({
        orgId: org!.id,
        installId: install!.id,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  return (
    <PageSection>
      <PageTitle title={`Versions | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/versions`,
            text: 'Versions',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          App config versions
        </Text>
        <Text variant="subtext" theme="neutral">
          History of app config updates applied to this install.
        </Text>
      </HeadingGroup>

      {isLoading && (
        <Card>
          <Text variant="subtext" theme="neutral">Loading...</Text>
        </Card>
      )}

      {!isLoading && (!versions || versions.length === 0) && (
        <Card>
          <EmptyState
            emptyTitle="No versions yet"
            emptyMessage="App config versions will appear here after the first sync."
            variant="table"
          />
        </Card>
      )}

      {versions && versions.length > 0 && (
        <div className="flex flex-col gap-3">
          {versions.map((version) => (
            <VersionCard key={version.id} version={version} />
          ))}
        </div>
      )}
    </PageSection>
  )
}

const VersionCard = ({ version }: { version: TInstallAppConfigVersion }) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const statusTheme = version.status?.status
    ? getStatusTheme(version.status.status)
    : 'neutral'

  const source = version.app_branch_run_id ? 'branch run' : 'sync'

  const { data: diff, isLoading: isDiffLoading } = useQuery({
    queryKey: [
      'install-app-config-version-diff',
      org?.id,
      install?.id,
      version.id,
    ],
    queryFn: () =>
      getInstallAppConfigVersionDiff({
        orgId: org!.id,
        installId: install!.id,
        versionId: version.id!,
      }),
    enabled: !!org?.id && !!install?.id && !!version.id,
    retry: 1,
  })

  const diffSummary = diff
    ? {
        added: diff.added?.length ?? 0,
        removed: diff.removed?.length ?? 0,
        changed: diff.changed?.length ?? 0,
      }
    : null

  return (
    <Expand
      id={`version-${version.id}`}
      className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-xl bg-white dark:bg-dark-grey-900 shadow-sm overflow-hidden"
      headerClassName="px-5 py-4"
      heading={
        <div className="flex items-center justify-between gap-4 w-full">
          <div className="flex items-center gap-3">
            <Badge
              size="sm"
              theme={
                statusTheme as
                  | 'success'
                  | 'error'
                  | 'info'
                  | 'neutral'
                  | 'warn'
              }
            >
              {version.status?.status ?? 'unknown'}
            </Badge>
            <Badge size="sm" theme="neutral">
              {source}
            </Badge>
            {version.created_at && (
              <Time
                variant="subtext"
                time={version.created_at}
                format="relative"
              />
            )}
          </div>
          {diffSummary && (
            <div className="flex items-center gap-2">
              {diffSummary.added > 0 && (
                <Badge size="sm" theme="success">
                  +{diffSummary.added}
                </Badge>
              )}
              {diffSummary.changed > 0 && (
                <Badge size="sm" theme="info">
                  ~{diffSummary.changed}
                </Badge>
              )}
              {diffSummary.removed > 0 && (
                <Badge size="sm" theme="error">
                  -{diffSummary.removed}
                </Badge>
              )}
            </div>
          )}
        </div>
      }
    >
      <div className="p-5 border-t border-cool-grey-100 dark:border-dark-grey-800">
        <div className="grid grid-cols-2 gap-3 mb-4">
          <div>
            <Text variant="subtext" theme="neutral">
              Old config
            </Text>
            <Text variant="subtext" family="mono">
              {version.old_app_config_id || 'none'}
            </Text>
          </div>
          <div>
            <Text variant="subtext" theme="neutral">
              New config
            </Text>
            <Text variant="subtext" family="mono">
              {version.new_app_config_id || 'none'}
            </Text>
          </div>
        </div>

        {version.metadata &&
          Object.keys(version.metadata).length > 0 && (
            <div className="flex flex-wrap gap-2 mb-4">
              {Object.entries(version.metadata).map(([key, value]) => (
                <Badge key={key} size="sm" theme="neutral">
                  {key}: {value}
                </Badge>
              ))}
            </div>
          )}

        {isDiffLoading && (
          <Text variant="subtext" theme="neutral">
            Loading diff...
          </Text>
        )}

        {diff && <DiffDetail diff={diff} />}

        {!isDiffLoading && !diff && (
          <Text variant="subtext" theme="neutral">
            No diff available
          </Text>
        )}
      </div>
    </Expand>
  )
}

const DiffDetail = ({ diff }: { diff: TInstallConfigDiff }) => {
  const hasComponents =
    (diff.added?.length ?? 0) +
      (diff.removed?.length ?? 0) +
      (diff.changed?.length ?? 0) >
    0

  if (!hasComponents && !diff.sandbox_changed && !diff.stack_changed) {
    return (
      <Text variant="subtext" theme="neutral">
        No changes detected
      </Text>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {diff.added?.map((entry) => (
        <DiffRow
          key={entry.component_id}
          op="added"
          entry={entry}
        />
      ))}
      {diff.changed?.map((entry) => (
        <DiffRow
          key={entry.component_id}
          op="changed"
          entry={entry}
        />
      ))}
      {diff.removed?.map((entry) => (
        <DiffRow
          key={entry.component_id}
          op="removed"
          entry={entry}
        />
      ))}

      {diff.sandbox_changed && (
        <div className="flex items-center gap-2">
          <Badge size="sm" theme="info">changed</Badge>
          <Text variant="subtext">Sandbox config</Text>
        </div>
      )}

      {diff.stack_changed && (
        <div className="flex items-center gap-2">
          <Badge size="sm" theme="info">changed</Badge>
          <Text variant="subtext">Stack config</Text>
        </div>
      )}
    </div>
  )
}

const DiffRow = ({
  op,
  entry,
}: {
  op: 'added' | 'changed' | 'removed'
  entry: { component_id: string; component_name?: string; component_type?: string }
}) => {
  const theme = op === 'added' ? 'success' : op === 'removed' ? 'error' : 'info'

  return (
    <div className="flex items-center gap-2">
      <Badge size="sm" theme={theme}>
        {op}
      </Badge>
      <Text variant="subtext">{entry.component_name || entry.component_id}</Text>
      {entry.component_type && (
        <Text variant="subtext" theme="neutral">
          {entry.component_type}
        </Text>
      )}
    </div>
  )
}
