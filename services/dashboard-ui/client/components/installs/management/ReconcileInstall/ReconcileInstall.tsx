import type { ReactNode } from 'react'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { SuccessfulBranchRuns } from '@/components/installs/InstallBranchBanner/SuccessfulBranchRuns'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type {
  TAPIError,
  TInstallOverviewCommit,
  TInstallReconciliation,
  TInstallReconciliationConfig,
} from '@/types'

interface IReconcileInstallModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  reconciliation?: TInstallReconciliation
  isLoading: boolean
  isPending: boolean
  error?: TAPIError | null
  repo?: string
  getRunHref: (run?: TInstallOverviewCommit) => string | undefined
  diff?: ReactNode
  onSubmit: () => void
}

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
  ...props
}: IReconcileInstallModal) => {
  const branchName = reconciliation?.app_branch_name
  const onBranch = branchName ? `the ${branchName} branch` : 'its app branch'
  const status = reconciliation?.status
  const canReconcile =
    status === 'behind' && !!reconciliation?.target?.app_config_id
  const blockers = reconciliation?.blockers ?? []
  const skippedCount = reconciliation?.skipped_run_count ?? 0

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
            <Icon variant="Loading" /> Reconciling install
          </span>
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
                Reconciling {installName} will move it from{' '}
                {configLabel(reconciliation.pinned)} to{' '}
                {configLabel(reconciliation.target)} on {onBranch}. The changes
                below are applied in one workflow.
              </Text>
            ) : null}

            {reconciliation.target ? (
              <div className="grid grid-cols-1 items-stretch gap-3 md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)]">
                <Card className="!gap-2 !p-4 min-w-0">
                  <Text variant="label" theme="neutral" weight="strong">
                    Selected
                  </Text>
                  <Text variant="body" weight="strong">
                    {configLabel(reconciliation.pinned)}
                  </Text>
                  <ConfigCommit
                    config={reconciliation.pinned}
                    repo={repo}
                    href={getRunHref(reconciliation.pinned?.commit)}
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
