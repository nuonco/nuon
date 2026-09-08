import type { TInstall, TInstallConfigSync } from '@/types'
import { latestBranchConfig } from '@/utils/branch-utils'
import { Badge } from '../../atoms/Badge'
import { Icon } from '../../atoms/Icon'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import { CommitSummary } from '../../molecules/CommitSummary'
import { OverviewCard, OverviewCardGrid } from '../../molecules/OverviewCard'

export interface IInstallOverviewCards {
  install?: TInstall
  latestSync?: TInstallConfigSync
  isLoading?: boolean
}

export const InstallOverviewCards = ({
  install,
  latestSync,
  isLoading = false,
}: IInstallOverviewCards) => {
  const driftCount = install?.drifted_objects?.length
  const branch = install?.app_branch
  const branchConfig = branch ? latestBranchConfig(branch) : undefined

  return (
    <OverviewCardGrid>
      <OverviewCard title="Health">
        {isLoading ? (
          <Status loading loadingWidth={10} />
        ) : install?.composite_health_status ? (
          <>
            <Status
              status={install.composite_health_status}
              description={install?.composite_health_status_description}
            />
            {install?.composite_health_status_description ? (
              <Text variant="caption" color="tertiary" lines={2}>
                {install.composite_health_status_description}
              </Text>
            ) : null}
          </>
        ) : (
          <>
            <Text color="tertiary">Not reported</Text>
            <Text variant="caption" color="tertiary">
              Health will appear after the first evaluation.
            </Text>
          </>
        )}
      </OverviewCard>

      <OverviewCard title="Drift">
        {isLoading ? (
          <>
            <Status loading loadingWidth={9} />
            <Text loading loadingWidth={18} variant="caption" />
          </>
        ) : driftCount === undefined ? (
          <>
            <Text color="tertiary">Not scanned</Text>
            <Text variant="caption" color="tertiary">
              Drift will appear after the first scan.
            </Text>
          </>
        ) : (
          <>
            <Status
              status={driftCount > 0 ? 'drifted' : 'no-drift'}
              label={driftCount > 0 ? 'Drift detected' : 'No drift'}
            />
            <Text variant="caption" color="tertiary">
              {driftCount === 0
                ? 'No resources have drifted.'
                : `${driftCount} resource${driftCount === 1 ? '' : 's'} drifted`}
            </Text>
          </>
        )}
      </OverviewCard>

      <OverviewCard title="Branch">
        <span className="flex items-center gap-2">
          <Icon variant="GitBranchIcon" size={18} />
          <Text
            family="mono"
            weight="medium"
            loading={isLoading}
            loadingWidth={12}
          >
            {branch?.name ?? '—'}
          </Text>
        </span>
        {isLoading ? (
          <Badge loading loadingWidth={8} />
        ) : branchConfig?.config_number ? (
          <Badge>Config v{branchConfig.config_number}</Badge>
        ) : null}
      </OverviewCard>

      <OverviewCard title="Last update">
        <CommitSummary
          commit={latestSync?.vcs_connection_commit}
          updatedAt={latestSync?.created_at}
          isLoading={isLoading}
        />
      </OverviewCard>
    </OverviewCardGrid>
  )
}
