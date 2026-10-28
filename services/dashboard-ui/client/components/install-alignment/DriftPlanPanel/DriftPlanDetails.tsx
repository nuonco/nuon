import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Divider } from '@/components/common/Divider'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type {
  TAPIError,
  TInstallAlignmentResource,
  TInstallOverviewCommit,
} from '@/types'
import { shortSha } from '../alignment-utils'

export interface IDriftPlanDetails {
  deploymentHref?: string
  onApproveAndResolve?: () => void
  plan: ReactNode
  planWorkflowHref?: string
  resolveDisabled?: boolean
  resolveError?: TAPIError | null
  resolving?: boolean
  resource: TInstallAlignmentResource
  selectedCommit?: TInstallOverviewCommit
  selectedCommitHref?: string
}

const PlanChange = ({
  count,
  label,
  theme,
}: {
  count?: number
  label: string
  theme: 'success' | 'warn' | 'error' | 'info'
}) =>
  count ? (
    <Badge size="sm" theme={theme}>
      {count} {label}
    </Badge>
  ) : null

export const DriftPlanDetails = ({
  deploymentHref,
  onApproveAndResolve,
  plan,
  planWorkflowHref,
  resolveDisabled = false,
  resolveError,
  resolving = false,
  resource,
  selectedCommit,
  selectedCommitHref,
}: IDriftPlanDetails) => {
  const summary = resource.plan
  const openDeployment = resource.open_deployment?.id
  const sha = shortSha(selectedCommit?.sha)

  return (
    <div
      className="flex flex-col gap-4 shrink-0"
      data-testid="drift-plan-details"
    >
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {sha ? (
          <LabeledValue label="Selected commit">
            {selectedCommitHref ? (
              <Link href={selectedCommitHref} className="!font-mono">
                {sha}
              </Link>
            ) : (
              <Text as="span" family="mono" variant="subtext">
                {sha}
              </Text>
            )}
          </LabeledValue>
        ) : null}
        {summary?.created_at ? (
          <LabeledValue label="Planned">
            <span className="flex flex-wrap items-center gap-2">
              <Time
                time={summary.created_at}
                format="relative"
                variant="subtext"
              />
              {planWorkflowHref ? (
                <Link href={planWorkflowHref} textVariant="subtext">
                  {summary.plan_only
                    ? 'View drift scan'
                    : 'View planning deployment'}
                </Link>
              ) : null}
            </span>
          </LabeledValue>
        ) : null}
        {summary ? (
          <LabeledValue label="Changes">
            <span className="flex flex-wrap items-center gap-1.5">
              <PlanChange
                count={summary.changes_create}
                label="create"
                theme="success"
              />
              <PlanChange
                count={summary.changes_update}
                label="update"
                theme="info"
              />
              <PlanChange
                count={summary.changes_replace}
                label="replace"
                theme="warn"
              />
              <PlanChange
                count={summary.changes_delete}
                label="delete"
                theme="error"
              />
              {!summary.changes_create &&
              !summary.changes_update &&
              !summary.changes_replace &&
              !summary.changes_delete ? (
                <Text variant="subtext" theme="neutral">
                  No resource changes counted
                </Text>
              ) : null}
            </span>
          </LabeledValue>
        ) : null}
      </div>

      <Divider dividerWord="Plan" />
      <div className="flex flex-col min-w-0 shrink-0">{plan}</div>

      {openDeployment && deploymentHref ? (
        <div className="flex flex-col sm:flex-row sm:items-center gap-3">
          <Button
            variant="primary"
            className="w-full sm:w-auto justify-center shrink-0"
            href={deploymentHref}
          >
            View deployment
          </Button>
          <Text variant="subtext" theme="neutral">
            A deployment is already open for this resource. View its progress.
          </Text>
        </div>
      ) : onApproveAndResolve ? (
        <div className="flex flex-col sm:flex-row sm:items-center gap-3">
          <Button
            variant="primary"
            className="w-full sm:w-auto justify-center shrink-0"
            onClick={onApproveAndResolve}
            disabled={resolving || resolveDisabled}
            tooltipProps={
              resolveDisabled
                ? {
                    tipContent: 'Load and review the plan before approving it',
                  }
                : undefined
            }
            aria-busy={resolving}
          >
            <Icon variant={resolving ? 'Loading' : 'CheckIcon'} size={16} />
            {resolving
              ? 'Starting deployment'
              : 'Approve plan and resolve drift'}
          </Button>
          <Text variant="subtext" theme="neutral">
            Starts a fresh deployment and reuses this approval only for an
            identical plan. Changed plans follow the install’s approval
            settings, and policy denials still block deployment.
          </Text>
        </div>
      ) : null}
      {resolveError ? (
        <Banner theme="error" role="alert">
          <Text variant="subtext">
            Drift resolution failed.{' '}
            {resolveError.description || resolveError.error}
          </Text>
        </Banner>
      ) : null}
    </div>
  )
}
