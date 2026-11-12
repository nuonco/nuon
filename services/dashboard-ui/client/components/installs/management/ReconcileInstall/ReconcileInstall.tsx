import type { ReactNode } from 'react'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { SuccessfulBranchRuns } from '@/components/installs/InstallBranchBanner/SuccessfulBranchRuns'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { Toggle } from '@/components/common/form/Toggle'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type {
  TAPIError,
  TInstallOverviewCommit,
  TInstallReconciliation,
  TInstallReconciliationConfig,
} from '@/types'
import {
  DEPLOY_PLAN_KIND_LABELS,
  DEPLOY_PLAN_REASON_LABELS,
  type TMockDeployPlanItem,
  type TMockInstallReconciliation,
} from '@/views/install/reconcile-mock-types'

interface IReconcileInstallModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  reconciliation?: TInstallReconciliation | TMockInstallReconciliation
  isLoading: boolean
  isPending: boolean
  error?: TAPIError | null
  repo?: string
  getRunHref: (run?: TInstallOverviewCommit) => string | undefined
  diff?: ReactNode
  onSubmit: () => void
  /** Proposed dry-run (plan only) option — presentational, not yet wired. */
  planOnly?: boolean
  onPlanOnlyChange?: (planOnly: boolean) => void
}

const DEPLOY_PLAN_KIND_ICONS: Record<TMockDeployPlanItem['kind'], TIconVariant> = {
  component: 'CubeIcon',
  image: 'CubeIcon',
  sandbox: 'ShippingContainerIcon',
  stack: 'StackIcon',
}

const DeployPlanList = ({ items }: { items: TMockDeployPlanItem[] }) => (
  <div className="flex flex-col gap-2">
    <Text variant="label" theme="neutral" weight="strong">
      Deploy plan
    </Text>
    <Card className="!p-0 !gap-0 divide-y" data-testid="reconcile-deploy-plan">
      {items
        .slice()
        .sort((a, b) => a.order - b.order)
        .map((item) => (
          <div
            key={item.id}
            className="flex flex-wrap items-center justify-between gap-2 p-3"
          >
            <span className="flex items-center gap-2 min-w-0">
              <Text
                variant="subtext"
                theme="neutral"
                family="mono"
                className="w-5 shrink-0 text-right"
              >
                {item.order}
              </Text>
              <Icon variant={DEPLOY_PLAN_KIND_ICONS[item.kind]} size={16} />
              <Text family="mono" weight="strong" className="break-all">
                {item.name}
              </Text>
              <Badge size="sm" theme="none">
                {DEPLOY_PLAN_KIND_LABELS[item.kind]}
              </Badge>
            </span>
            <span className="flex items-center gap-2">
              <Badge size="sm" theme="warn">
                {DEPLOY_PLAN_REASON_LABELS[item.reason]}
              </Badge>
            </span>
            <Text variant="subtext" theme="neutral" className="basis-full">
              {item.reason_message}
            </Text>
          </div>
        ))}
    </Card>
  </div>
)

const configLabel = (config?: TInstallReconciliationConfig) =>
  config?.version ? `config v${config.version}` : 'no config'

const ConfigCommit = ({
  config,
  repo,
  href,
}: {
  config?: TInstallReconciliationConfig
  repo?: string
  href?: string
}) =>
  config?.commit?.sha ? (
    <BranchRunCommit
      displayVariant="inline"
      href={href}
      sha={config.commit.sha}
      message={config.commit.message || 'App configuration update'}
      author={config.commit.author}
      createdAt={config.commit.created_at}
      repo={repo}
      showStatus={false}
    />
  ) : null

export const ReconcileInstallModal = ({
  installName,
  reconciliation,
  isLoading,
  isPending,
  error,
  repo,
  getRunHref,
  diff,
  onSubmit,
  planOnly,
  onPlanOnlyChange,
  ...props
}: IReconcileInstallModal) => {
  const branchName = reconciliation?.app_branch_name
  const onBranch = branchName ? `the ${branchName} branch` : 'its app branch'
  const status = reconciliation?.status
  const outOfSync =
    status === 'out_of_sync' ||
    !!(reconciliation as TMockInstallReconciliation | undefined)?.out_of_sync
  const canReconcile =
    (status === 'behind' || status === 'out_of_sync') &&
    !!reconciliation?.target?.app_config_id
  const blockers = reconciliation?.blockers ?? []
  const skippedCount = reconciliation?.skipped_run_count ?? 0
  const deployPlan = (reconciliation as TMockInstallReconciliation | undefined)
    ?.deploy_plan

  return (
    <Modal
      size="xl"
      heading={
        <Text flex className="gap-3" variant="h3" weight="strong">
          <Icon variant="ArrowsClockwiseIcon" size="24" />
          Reconcile install
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />{' '}
            {planOnly ? 'Planning reconcile' : 'Reconciling install'}
          </span>
        ) : planOnly ? (
          'Plan reconcile'
        ) : (
          'Reconcile install'
        ),
        disabled: !canReconcile || isPending,
        onClick: onSubmit,
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        disabled: isPending,
        onClick: () => props.onClose?.(),
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to reconcile this install.'}
          </Banner>
        ) : null}

        {isLoading ? (
          <div className="flex justify-center py-8">
            <Loading variant="large" />
          </div>
        ) : null}

        {!isLoading && !reconciliation ? (
          <Banner theme="error">
            Unable to load reconciliation details. Try again.
          </Banner>
        ) : null}

        {reconciliation ? (
          <>
            {status === 'current' ? (
              <Banner theme="info">
                {installName} already runs the latest config on {onBranch}.
              </Banner>
            ) : null}

            {status === 'in_progress' ? (
              <Banner theme="info">
                A workflow is already reconciling {installName} onto {onBranch}.
              </Banner>
            ) : null}

            {status === 'blocked'
              ? blockers.map((blocker) => (
                  <Banner key={blocker.code} theme="warn">
                    <strong>Cannot reconcile:</strong> {blocker.message}
                  </Banner>
                ))
              : null}

            {canReconcile ? (
              <Text variant="body">
                {outOfSync ? (
                  <>
                    {installName} isn’t fully running{' '}
                    {configLabel(reconciliation.target)} on {onBranch}.
                    Reconciling will redeploy the components below in order.
                  </>
                ) : (
                  <>
                    Reconciling {installName} will move it from{' '}
                    {configLabel(reconciliation.applied)} to{' '}
                    {configLabel(reconciliation.target)} on {onBranch}. The
                    changes below are applied in one workflow.
                  </>
                )}
              </Text>
            ) : null}

            {canReconcile && onPlanOnlyChange ? (
              <Toggle
                checked={!!planOnly}
                onChange={onPlanOnlyChange}
                label="Dry run (plan only)"
                description="Preview the deploy plan without applying any changes."
              />
            ) : null}

            {canReconcile && deployPlan?.length ? (
              <DeployPlanList items={deployPlan} />
            ) : null}

            {reconciliation.target ? (
              <div className="grid grid-cols-1 items-stretch gap-3 md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)]">
                <Card className="!gap-2 !p-4 min-w-0">
                  <Text variant="label" theme="neutral" weight="strong">
                    Applied
                  </Text>
                  <Text variant="body" weight="strong">
                    {configLabel(reconciliation.applied)}
                  </Text>
                  <ConfigCommit
                    config={reconciliation.applied}
                    repo={repo}
                    href={getRunHref(reconciliation.applied?.commit)}
                  />
                </Card>
                <div className="hidden items-center justify-center px-1 md:flex">
                  <Icon
                    variant="ArrowRightIcon"
                    size={20}
                    className="text-cool-grey-400"
                  />
                </div>
                <Card className="!gap-2 !p-4 min-w-0">
                  <Text variant="label" theme="neutral" weight="strong">
                    Target
                  </Text>
                  <Text variant="body" weight="strong">
                    {configLabel(reconciliation.target)}
                  </Text>
                  <ConfigCommit
                    config={reconciliation.target}
                    repo={repo}
                    href={getRunHref(reconciliation.target.commit)}
                  />
                </Card>
              </div>
            ) : null}

            {skippedCount > 0 ? (
              <Expand
                id="reconcile-skipped-runs"
                heading={
                  <Text weight="strong">
                    {skippedCount === 1
                      ? 'Behind 1 branch run. Reconciling applies this change.'
                      : `Behind ${skippedCount} branch runs. Reconciling applies these changes.`}
                  </Text>
                }
              >
                <div className="p-2">
                  <SuccessfulBranchRuns
                    branchName={reconciliation.app_branch_name}
                    runs={reconciliation.skipped_runs ?? []}
                    runsBehind={skippedCount}
                    repo={repo}
                    getRunHref={getRunHref}
                  />
                </div>
              </Expand>
            ) : null}

            {diff}
          </>
        ) : null}
      </div>
    </Modal>
  )
}
