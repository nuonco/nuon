import type { TInstall, TVCSCommit } from '@/types'
import { Icon } from '../../atoms/Icon'
import { Link } from '../../atoms/Link'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import { CommitSummary } from '../../molecules/CommitSummary'
import { OverviewCard, OverviewCardGrid } from '../../molecules/OverviewCard'
import { installStatusFacets } from '../../../utils/install-details'

export interface IInstallBranchUpdate {
  runId?: string
  runHref?: string
  branchName?: string
  status?: string
  commit?: TVCSCommit
  updatedAt?: string
}

export interface IInstallOverviewCards {
  install?: TInstall
  lastBranchUpdate?: IInstallBranchUpdate
  loading?: boolean
}

const SERVICE_LABELS: Record<string, string> = {
  runner: 'Runner',
  sandbox: 'Sandbox',
  components: 'Components',
}

export const InstallOverviewCards = ({
  install,
  lastBranchUpdate,
  loading = false,
}: IInstallOverviewCards) => {
  const driftCount = install?.drifted_objects?.length
  const lifecycle = install?.lifecycle_phase?.phase
  const services = installStatusFacets(install).filter((facet) =>
    ['runner', 'sandbox', 'components'].includes(facet.id)
  )

  return (
    <OverviewCardGrid columns={3}>
      <OverviewCard title="Install status">
        {loading ? (
          <>
            <Status loading loadingWidth={10} />
            <Status loading loadingWidth={12} />
          </>
        ) : (
          <>
            <Status
              status={lifecycle ?? 'unknown'}
              description={install?.lifecycle_phase?.description}
            />
            {install?.composite_health_status ? (
              <Status
                status={install.composite_health_status}
                description={install?.composite_health_status_description}
              />
            ) : (
              <Text color="tertiary">Health not reported</Text>
            )}
            <Text variant="caption" color="tertiary">
              {driftCount === undefined
                ? 'Drift not scanned'
                : driftCount === 0
                  ? 'No drift detected'
                  : `${driftCount} resource${driftCount === 1 ? '' : 's'} drifted`}
            </Text>
          </>
        )}
      </OverviewCard>

      <OverviewCard
        title="Expected update"
        footer={
          !loading && lastBranchUpdate?.runHref ? (
            <Link href={lastBranchUpdate.runHref}>View activity</Link>
          ) : undefined
        }
      >
        <span className="flex flex-wrap items-center justify-between gap-2">
          <span className="flex items-center gap-2">
            <Icon variant="GitBranchIcon" size={18} />
            <Text
              family="mono"
              weight="medium"
              loading={loading}
              loadingWidth={12}
            >
              {lastBranchUpdate?.branchName ??
                install?.app_branch?.name ??
                '—'}
            </Text>
          </span>
          {loading ? (
            <Status loading />
          ) : lastBranchUpdate?.status ? (
            <Status status={lastBranchUpdate.status} variant="inline" />
          ) : null}
        </span>
        <CommitSummary
          commit={lastBranchUpdate?.commit}
          updatedAt={lastBranchUpdate?.updatedAt}
          loading={loading}
        />
        {!loading && lastBranchUpdate?.runId ? (
          <Text variant="label" family="mono" color="tertiary">
            {lastBranchUpdate.runId}
          </Text>
        ) : null}
      </OverviewCard>

      <OverviewCard title="Current services">
        {loading ? (
          <>
            <Status loading loadingWidth={10} />
            <Status loading loadingWidth={10} />
            <Status loading loadingWidth={10} />
          </>
        ) : (
          <div className="flex flex-col gap-3">
            {services.map((service) => (
              <div
                key={service.id}
                className="flex items-center justify-between gap-3"
              >
                <Text variant="caption" color="secondary">
                  {SERVICE_LABELS[service.id] ?? service.id}
                </Text>
                <Status
                  status={service.status}
                  description={service.description}
                  variant="inline"
                />
              </div>
            ))}
          </div>
        )}
      </OverviewCard>
    </OverviewCardGrid>
  )
}
