import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TInstallAlignmentResource, TInstallOverviewCommit } from '@/types'
import { shortSha } from '../alignment-utils'

export interface IDriftPlanDetails {
  deploymentHref?: string
  plan: ReactNode
  planWorkflowHref?: string
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
  plan,
  planWorkflowHref,
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
      <div className="flex flex-wrap gap-x-6 gap-y-3">
        {sha ? (
          <LabeledValue label="Selected commit">
            {selectedCommitHref ? (
              <Link href={selectedCommitHref}>
                <Text as="span" family="mono" variant="subtext">
                  {sha}
                </Text>
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

      <div className="flex flex-col min-w-0 shrink-0">{plan}</div>

      {openDeployment && deploymentHref ? (
        <div className="flex flex-col sm:flex-row sm:items-center gap-3">
          <Button
            size="lg"
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
      ) : null}
    </div>
  )
}
