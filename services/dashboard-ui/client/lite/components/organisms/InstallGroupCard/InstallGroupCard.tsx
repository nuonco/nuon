import { Card } from '../../atoms/Card'
import { Badge } from '../../atoms/Badge'
import { Link } from '../../atoms/Link'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import { ConfigItem } from '../../molecules/ConfigItem'
import { LabelSelectorSummary } from '../../molecules/LabelSelectorSummary'
import type {
  IDeploymentPlanInstall,
  IDeploymentPlanStage,
} from '../../../utils/deployment-plan'

export const MAX_INLINE_INSTALLS = 5

export interface IInstallGroupCard {
  stage?: IDeploymentPlanStage
  labelColors?: Record<string, string>
  groupHref?: string
  onInstallSelect?: (install: IDeploymentPlanInstall) => void
  loading?: boolean
}

const membershipText = (stage: IDeploymentPlanStage) => {
  if (stage.membership === 'all_installs') {
    return 'Every install on this branch'
  }
  if (stage.membership === 'label_selector') {
    return 'Installs matching these labels'
  }
  return `A fixed list of ${stage.installs.length} ${
    stage.installs.length === 1 ? 'install' : 'installs'
  }`
}

const installCountText = (count: number) =>
  `${count} ${count === 1 ? 'install' : 'installs'}`

export const InstallGroupCard = ({
  stage,
  labelColors,
  groupHref,
  onInstallSelect,
  loading = false,
}: IInstallGroupCard) => {
  const visibleInstalls = stage?.installs.slice(0, MAX_INLINE_INSTALLS) ?? []
  const awaitingApproval = stage?.status === 'approval-awaiting'

  return (
    <Card as="article" className="flex flex-col gap-4">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <span className="flex min-w-0 items-center gap-3">
          <Badge tone="accent">
            {loading ? 'Stage' : `Stage ${stage?.stage ?? '—'}`}
          </Badge>
          <Text
            as="h3"
            variant="heading"
            className="min-w-0 truncate"
            loading={loading}
            loadingWidth={16}
          >
            {stage?.name ?? '—'}
          </Text>
        </span>
        {loading ? (
          <Status loading />
        ) : stage?.status ? (
          <Status
            status={stage.status}
            label={awaitingApproval ? 'Waiting for approval' : undefined}
            theme={awaitingApproval ? 'warn' : undefined}
          />
        ) : null}
      </header>

      <div className="flex flex-col gap-2">
        <Text
          color="secondary"
          loading={loading}
          loadingWidth={24}
        >
          {stage ? membershipText(stage) : '—'}
        </Text>
        {stage?.membership === 'label_selector' ? (
          <>
            <LabelSelectorSummary
              selector={stage.selector}
              labelColors={labelColors}
            />
            <Text variant="caption" color="tertiary">
              Membership is resolved at rollout time. These installs match now.
            </Text>
          </>
        ) : null}
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <Text
          weight="semibold"
          loading={loading}
          loadingWidth={10}
        >
          {installCountText(stage?.totalInstalls ?? 0)}
        </Text>
        {stage?.completedInstalls !== undefined ? (
          <Text variant="caption" color="secondary">
            {stage.completedInstalls} done
          </Text>
        ) : null}
        {(stage?.failedInstalls ?? 0) > 0 ? (
          <Text variant="caption" className="text-status-error">
            {stage?.failedInstalls} failed
          </Text>
        ) : null}
      </div>

      {loading ? (
        <div className="flex flex-col gap-1">
          <ConfigItem loading />
          <ConfigItem loading />
        </div>
      ) : visibleInstalls.length > 0 ? (
        <div className="flex flex-col gap-1 border-t pt-2">
          {visibleInstalls.map((install) => (
            <ConfigItem
              key={install.id}
              icon="ShippingContainerIcon"
              name={install.name ?? install.id}
              id={install.id}
              onClick={() => onInstallSelect?.(install)}
              disabled={!onInstallSelect}
            />
          ))}
        </div>
      ) : (
        <Text variant="caption" color="tertiary">
          No installs match this group.
        </Text>
      )}

      <footer className="flex flex-wrap items-center justify-between gap-3 border-t pt-3">
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
        {stage && stage.totalInstalls > 0 && groupHref ? (
          <Link href={groupHref} variant="caption">
            View all {installCountText(stage.totalInstalls)}
          </Link>
        ) : null}
      </footer>
    </Card>
  )
}
