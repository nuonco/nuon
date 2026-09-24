import type { ReactNode } from 'react'
import type { TAppBranch } from '@/types'
import { Card } from '../../atoms/Card'
import { Icon } from '../../atoms/Icon'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import type {
  IDeploymentPlanInstall,
  IDeploymentPlanStage,
} from '../../../utils/deployment-plan'
import { InstallGroupCard } from '../InstallGroupCard'

export interface IDeploymentPlanStages {
  branch?: TAppBranch
  stages?: IDeploymentPlanStage[]
  labelColors?: Record<string, string>
  groupHref?: (stage: IDeploymentPlanStage) => string | undefined
  onInstallSelect?: (install: IDeploymentPlanInstall) => void
  renderStageCard?: (stage: IDeploymentPlanStage) => ReactNode
  loading?: boolean
  error?: unknown
}

const EmptyPlan = ({ error }: { error?: unknown }) => (
  <Card className="flex flex-col gap-2">
    <Text as="h3" variant="heading">
      {error ? 'Deployment plan failed to load' : 'No deployment plan configured'}
    </Text>
    <Text color="secondary">
      {error
        ? 'The deployment stages could not be loaded. Try again.'
        : 'Install groups will appear here once this branch has a deployment plan.'}
    </Text>
  </Card>
)

const BranchSummary = ({
  branch,
  loading,
}: {
  branch?: TAppBranch
  loading: boolean
}) => {
  const run = branch?.latest_run
  const commit = run?.vcs_connection_commit?.sha ?? run?.head_sha

  return (
    <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
      <span className="flex min-w-0 items-center gap-2">
        <Icon variant="GitBranchIcon" size={18} />
        <Text
          family="mono"
          weight="semibold"
          loading={loading}
          loadingWidth={12}
        >
          {branch?.name ?? '—'}
        </Text>
      </span>
      {loading ? (
        <Text variant="caption" loading loadingWidth={8} />
      ) : commit ? (
        <span className="flex items-center gap-1.5">
          <Icon variant="GitCommitIcon" size={14} />
          <Text variant="caption" family="mono" color="tertiary">
            {commit.slice(0, 7)}
          </Text>
        </span>
      ) : (
        <Text variant="caption" color="tertiary">
          No rollouts yet
        </Text>
      )}
      {loading ? (
        <Status loading />
      ) : run?.status ? (
        <Status
          status={run.status}
          label={run.awaiting_approval ? 'Waiting for approval' : undefined}
          theme={run.awaiting_approval ? 'warn' : undefined}
        />
      ) : null}
    </div>
  )
}

export const DeploymentPlanStages = ({
  branch,
  stages,
  labelColors,
  groupHref,
  onInstallSelect,
  renderStageCard,
  loading = false,
  error,
}: IDeploymentPlanStages) => {
  const visibleStages = loading
    ? Array.from({ length: 3 }, () => undefined)
    : (stages ?? [])

  return (
    <section className="flex flex-col gap-4" aria-label="Deployment plan">
      <header className="flex flex-col gap-2">
        <Text as="h2" variant="heading">
          Deployment plan
        </Text>
        <BranchSummary branch={branch} loading={loading} />
      </header>
      {error || (!loading && visibleStages.length === 0) ? (
        <EmptyPlan error={error} />
      ) : (
        <div>
          {visibleStages.map((stage, index) => (
            <div key={stage?.id ?? `loading-${index}`}>
              {index > 0 ? (
                <div
                  aria-hidden
                  className="ml-6 h-4 border-l"
                />
              ) : null}
              {stage && renderStageCard ? (
                renderStageCard(stage)
              ) : (
                <InstallGroupCard
                  stage={stage}
                  labelColors={labelColors}
                  groupHref={stage ? groupHref?.(stage) : undefined}
                  onInstallSelect={onInstallSelect}
                  loading={loading}
                />
              )}
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
