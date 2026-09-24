import type { TInstallGroupRun } from '@/types'
import { Badge } from '../../atoms/Badge'
import { Card } from '../../atoms/Card'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import { CloudRegion } from '../../molecules/CloudRegion'
import { ConfigItem } from '../../molecules/ConfigItem'
import { LabelSelectorSummary } from '../../molecules/LabelSelectorSummary'
import { Pagination } from '../../molecules/Pagination'
import { Time } from '../../molecules/Time'
import type {
  IDeploymentPlanInstall,
  IDeploymentPlanStage,
} from '../../../utils/deployment-plan'
import {
  installCloudLocation,
  installStatusFacets,
} from '../../../utils/install-details'
import { Panel } from '../surfaces'

export interface IInstallGroupPanel {
  stage?: IDeploymentPlanStage
  run?: TInstallGroupRun
  labelColors?: Record<string, string>
  offset: number
  pageSize: number
  onOffsetChange: (offset: number) => void
  onInstallSelect?: (install: IDeploymentPlanInstall) => void
  loading?: boolean
  error?: unknown
}

const membershipText = (stage: IDeploymentPlanStage) => {
  if (stage.membership === 'all_installs') {
    return 'Every install on this branch, after earlier groups'
  }
  if (stage.membership === 'label_selector') {
    return 'Installs matching these labels'
  }
  return `A fixed list of ${stage.installs.length} ${
    stage.installs.length === 1 ? 'install' : 'installs'
  }`
}

const InstallStatuses = ({
  install,
}: {
  install: IDeploymentPlanInstall
}) => (
  <>
    {installStatusFacets(install).map((facet) => (
      <Status
        key={facet.id}
        status={facet.status}
        icon={facet.icon}
        label={facet.title}
        description={facet.description}
        variant="icon"
      />
    ))}
  </>
)

const InstallLocation = ({
  install,
}: {
  install: IDeploymentPlanInstall
}) => {
  const location = installCloudLocation(install)
  if (!location.region && !location.location) return null

  return <CloudRegion {...location} lines={1} />
}

export const InstallGroupPanel = ({
  stage,
  run,
  labelColors,
  offset,
  pageSize,
  onOffsetChange,
  onInstallSelect,
  loading = false,
  error,
}: IInstallGroupPanel) => {
  const installs = stage?.installs ?? []
  const visibleInstalls = installs.slice(offset, offset + pageSize)
  const awaitingApproval = stage?.status === 'approval-awaiting'

  return (
    <Panel heading={stage?.name ?? 'Install group'} defaultSize="half">
      {error ? (
        <Card className="flex flex-col gap-2">
          <Text as="h3" variant="heading">
            Install group failed to load
          </Text>
          <Text color="secondary">
            This group could not be loaded. Try again.
          </Text>
        </Card>
      ) : (
        <>
          <Card className="flex flex-col gap-3">
            <Text as="h3" variant="caption" color="secondary" weight="medium">
              Membership
            </Text>
            <Text loading={loading} loadingWidth={24}>
              {stage ? membershipText(stage) : '—'}
            </Text>
            {stage?.membership === 'label_selector' ? (
              <>
                <LabelSelectorSummary
                  selector={stage.selector}
                  labelColors={labelColors}
                />
                <Text variant="caption" color="tertiary">
                  Membership is resolved at rollout time. These installs match
                  now.
                </Text>
              </>
            ) : null}
            <span className="flex flex-wrap items-center gap-2">
              {stage?.maxParallel !== undefined ? (
                <Badge>Max parallel {stage.maxParallel}</Badge>
              ) : null}
              {stage?.autoApproveOnPoliciesPassing !== undefined ? (
                <Badge>
                  {stage.autoApproveOnPoliciesPassing
                    ? 'Auto-approve'
                    : 'Manual approval'}
                </Badge>
              ) : null}
            </span>
          </Card>

          <Card className="flex flex-col gap-3">
            <Text as="h3" variant="caption" color="secondary" weight="medium">
              Latest rollout
            </Text>
            {loading ? (
              <Status loading />
            ) : stage?.status ? (
              <Status
                status={stage.status}
                label={awaitingApproval ? 'Waiting for approval' : undefined}
                theme={awaitingApproval ? 'warn' : undefined}
              />
            ) : (
              <Text color="tertiary">Not rolled out yet</Text>
            )}
            {run ? (
              <>
                <Text color="secondary">
                  {run.completed_installs ?? 0} done ·{' '}
                  {run.failed_installs ?? 0} failed ·{' '}
                  {run.total_installs ?? 0} total
                </Text>
                <Time value={run.updated_at ?? run.created_at} format="relative" />
              </>
            ) : null}
          </Card>

          <Card className="flex flex-col gap-3">
            <span className="flex items-baseline justify-between gap-3">
              <Text as="h3" variant="caption" color="secondary" weight="medium">
                Installs
              </Text>
              <Text variant="caption" color="tertiary">
                {stage?.totalInstalls ?? installs.length} total
              </Text>
            </span>
            {loading ? (
              <div className="flex flex-col gap-1">
                <ConfigItem loading />
                <ConfigItem loading />
                <ConfigItem loading />
              </div>
            ) : visibleInstalls.length > 0 ? (
              <div className="flex flex-col gap-1">
                {visibleInstalls.map((install) => (
                  <ConfigItem
                    key={install.id}
                    name={install.name ?? install.id}
                    id={install.id}
                    status={<InstallStatuses install={install} />}
                    metadata={<InstallLocation install={install} />}
                    onClick={() => onInstallSelect?.(install)}
                    disabled={!onInstallSelect}
                  />
                ))}
              </div>
            ) : (
              <Text color="tertiary">No installs match this group.</Text>
            )}
            {installs.length > pageSize ? (
              <Pagination
                label="Group installs pagination"
                offset={offset}
                pageSize={pageSize}
                hasNext={offset + pageSize < installs.length}
                onOffsetChange={onOffsetChange}
              />
            ) : null}
          </Card>
        </>
      )}
    </Panel>
  )
}
