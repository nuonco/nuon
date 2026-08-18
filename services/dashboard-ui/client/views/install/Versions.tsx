import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { AppConfigDiff } from '@/components/branches/AppConfigDiff'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallAppConfigVersions } from '@/lib'
import type { TInstallAppConfigVersion } from '@/types'

export const Versions = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: versions, isLoading } = useQuery({
    placeholderData: keepPreviousData,
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
          App branch runs
        </Text>
        <Text variant="subtext" theme="neutral">
          History of app config updates applied to this install.
        </Text>
        {install?.app_branch_id && (
          <div className="flex items-center gap-2 mt-1">
            <Badge size="sm" theme="info">
              {install.app_branch?.name || 'branch'}
            </Badge>
            <Link
              href={`/${org?.id}/apps/${install?.app_id}/branches/${install?.app_branch_id}`}
              className="text-xs"
            >
              View branch
            </Link>
          </div>
        )}
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

  const source = version.metadata?.triggered_by || (version.app_branch_run_id ? 'app-branch' : 'sync')

  return (
    <Expand
      id={`version-${version.id}`}
      className="border rounded-xl bg-white dark:bg-dark-grey-900 shadow-sm overflow-hidden"
      headerClassName="px-5 py-4"
      heading={
        <div className="flex items-center gap-3 w-full">
          <Status status={version.status?.status || 'unknown'} variant="badge" />
          <Badge size="sm" theme="neutral">
            {source}
          </Badge>
          {version.created_at && (
            <Time
              variant="subtext"
              theme="neutral"
              time={version.created_at}
              format="relative"
            />
          )}
        </div>
      }
    >
      <div className="p-5 border-t border-cool-grey-100 dark:border-dark-grey-800">
        {version.app_branch_run?.vcs_connection_commit && (
          <div className="flex items-center gap-2 mb-4 px-3 py-2.5 rounded-lg bg-cool-grey-50 dark:bg-dark-grey-800">
            {version.app_branch_run?.app_branch?.name && (
              <Badge size="sm" theme="info">
                {version.app_branch_run.app_branch.name}
              </Badge>
            )}
            <Text variant="subtext" theme="neutral" className="font-mono">
              {version.app_branch_run.vcs_connection_commit.sha?.slice(0, 7)}
            </Text>
            <Text variant="subtext" theme="neutral" className="truncate">
              {version.app_branch_run.vcs_connection_commit.message}
            </Text>
            {version.app_branch_run?.pr_number && (
              <Badge size="sm" theme="neutral">
                PR #{version.app_branch_run.pr_number}
              </Badge>
            )}
          </div>
        )}
        <div className="grid grid-cols-2 gap-3 mb-4">
          <LabeledValue label="Old config">
            {version.old_app_config_id ? (
              <ID>{version.old_app_config_id}</ID>
            ) : (
              <Text variant="subtext" theme="neutral">
                none
              </Text>
            )}
          </LabeledValue>
          <LabeledValue label="New config">
            {version.new_app_config_id ? (
              <ID>{version.new_app_config_id}</ID>
            ) : (
              <Text variant="subtext" theme="neutral">
                none
              </Text>
            )}
          </LabeledValue>
        </div>

        <div className="flex flex-wrap items-center gap-2 mb-4">
          {version.metadata &&
            Object.entries(version.metadata).map(([key, value]) => (
              <Badge key={key} size="sm" theme="neutral">
                {key}: {value}
              </Badge>
            ))}
          {version.app_branch_run?.workflow_id && version.app_branch_run?.app_branch?.id && org?.id && (
            <Link
              href={`/${org.id}/apps/${install?.app_id}/branches/${version.app_branch_run?.app_branch?.id}/runs/${version.app_branch_run.workflow_id}`}
              className="text-xs"
            >
              View branch run
            </Link>
          )}
          {version.workflow_id && org?.id && install?.id && (
            <Link
              href={`/${org.id}/installs/${install.id}/workflows/${version.workflow_id}`}
              className="text-xs"
            >
              View workflow
            </Link>
          )}
        </div>

        {version.new_app_config_id && install?.app_id && (
          <AppConfigDiff
            appConfigId={version.new_app_config_id}
            oldConfigId={version.old_app_config_id}
            appId={install.app_id}
          />
        )}
      </div>
    </Expand>
  )
}

