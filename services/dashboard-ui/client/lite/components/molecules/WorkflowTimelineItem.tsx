import type { HTMLAttributes } from 'react'
import type { TAppBranchRun, TWorkflow } from '@/types/ctl-api.types'
import type { TStatusTheme } from '@/utils/status-utils'
import { WORKFLOW_BADGE_MAP, isServiceAccount } from '@/utils/workflow-utils'
import { Badge } from '../atoms/Badge'
import { Status } from '../atoms/Status'
import { Text } from '../atoms/Text'
import { Link } from '../atoms/Link'
import { Tooltip } from '../atoms/Tooltip'
import { Duration } from './Duration'
import { ID } from './ID'
import { TimelineItem } from './TimelineItem'

export interface IWorkflowTimelineItem
  extends Omit<HTMLAttributes<HTMLLIElement>, 'title'> {
  workflow: TWorkflow
  href?: string
  pendingApprovalIds?: ReadonlySet<string>
  driftedWorkflowIds?: ReadonlySet<string>
  loading?: boolean
}

type TCaptionBadge = {
  id: string
  label: string
  theme?: TStatusTheme
}

const DASHBOARD_THEMES: Record<string, TStatusTheme> = {
  success: 'success',
  error: 'error',
  warn: 'warn',
  info: 'info',
  neutral: 'neutral',
  brand: 'brand',
}

const CRON_TYPES = ['drift_run', 'drift_run_reprovision_sandbox']

const AUTO_APPROVE_LABELS: Record<string, string> = {
  'approve-workflow': 'auto-approve (workflow)',
  'install-config': 'auto-approve (config)',
}

const IN_FLIGHT_LABELS: Record<string, string> = {
  'in-progress': 'In progress',
  planning: 'Planning',
  applying: 'Applying',
  'checking-plan': 'Checking plan',
  retrying: 'Retrying',
  pending: 'Pending',
  queued: 'Queued',
}

const statusChip = (workflow: TWorkflow) => {
  const status = workflow?.status?.status
  if (!status) return undefined

  const badge = WORKFLOW_BADGE_MAP[status]
  if (badge) {
    return {
      label: badge.children,
      theme: badge.theme ? DASHBOARD_THEMES[badge.theme] : undefined,
    }
  }

  const inFlight = IN_FLIGHT_LABELS[status]
  return inFlight ? { label: inFlight, theme: 'info' as TStatusTheme } : undefined
}

const branchRunOf = (workflow: TWorkflow): TAppBranchRun | undefined =>
  workflow?.app_branch_runs?.at(0)

const isPreviewRun = (workflow: TWorkflow): boolean => {
  const branchRun = branchRunOf(workflow)
  return (
    !!branchRun?.preview ||
    !!branchRun?.plan_only ||
    branchRun?.run_type === 'git-preview-run'
  )
}

const previewSourceLabel = (workflow: TWorkflow): string | undefined => {
  const branchRun = branchRunOf(workflow)
  const preview = branchRun?.preview

  if (preview?.source === 'pr' && branchRun?.pr_number != null) {
    return `PR #${branchRun.pr_number}`
  }
  if (preview?.git_ref && preview.git_ref !== branchRun?.app_branch?.name) {
    return `git ref: ${preview.git_ref}`
  }
  if (preview?.source === 'branch' && branchRun?.base_branch) {
    return `git ref: ${branchRun.base_branch}`
  }
  return undefined
}

const isBranchRun = (workflow: TWorkflow): boolean =>
  workflow?.owner_type === 'app_branches'

const captionBadges = (
  workflow: TWorkflow,
  driftedWorkflowIds?: ReadonlySet<string>
): TCaptionBadge[] => {
  const badges: TCaptionBadge[] = []
  const branchRun = branchRunOf(workflow)

  if (isBranchRun(workflow) && isPreviewRun(workflow)) {
    badges.push({ id: 'preview', label: 'preview' })

    const mode = branchRun?.preview?.mode
    if (mode && mode !== 'plan-only') badges.push({ id: 'mode', label: mode })

    const source = previewSourceLabel(workflow)
    if (source) badges.push({ id: 'source', label: source })

    const installName = branchRun?.preview?.install_name
    if (installName)
      badges.push({ id: 'install', label: `install: ${installName}` })
  } else if (workflow?.plan_only) {
    if (isBranchRun(workflow)) {
      badges.push({ id: 'preview', label: 'preview' })
    } else {
      badges.push({ id: 'drift-scan', label: 'drift scan' })
      if (workflow?.id && driftedWorkflowIds?.has(workflow.id)) {
        badges.push({
          id: 'drift-detected',
          label: 'drift detected',
          theme: 'warn',
        })
      }
    }
  }

  if (isBranchRun(workflow) && branchRun?.event_type === 'manual') {
    badges.push({ id: 'manual', label: 'manual' })
  }

  if (workflow?.type && CRON_TYPES.includes(workflow.type)) {
    badges.push({ id: 'cron', label: 'cron scheduled' })
  }

  const approvalType = workflow?.metadata?.approval_type
  if (workflow?.approval_option === 'approve-all' && approvalType) {
    badges.push({
      id: 'auto-approve',
      label: AUTO_APPROVE_LABELS[approvalType] ?? 'auto-approve',
    })
  }

  return badges
}

const TriggeredBy = ({ workflow }: { workflow: TWorkflow }) => {
  const account = workflow?.created_by
  const email = account?.email
  if (!email) return null

  if (isServiceAccount(account)) {
    return (
      <Tooltip content={email}>
        <Text variant="caption" color="tertiary">
          a service account
        </Text>
      </Tooltip>
    )
  }

  return (
    <Text variant="caption" color="tertiary" lines={1}>
      {email}
    </Text>
  )
}

export const WorkflowTimelineItem = ({
  workflow,
  href,
  pendingApprovalIds,
  driftedWorkflowIds,
  loading = false,
  ...props
}: IWorkflowTimelineItem) => {
  if (loading) return <TimelineItem title="" loading {...props} />

  const chip = statusChip(workflow)
  const status = workflow?.status?.status
  const pendingApproval =
    workflow?.approval_option === 'prompt' &&
    status !== 'approval-awaiting' &&
    !!workflow?.id &&
    !!pendingApprovalIds?.has(workflow.id)

  const title = workflow?.name || workflow?.type || workflow?.id || 'Run'
  const badges = captionBadges(workflow, driftedWorkflowIds)

  return (
    <TimelineItem
      title={
        <span className="flex min-w-0 flex-wrap items-center gap-2">
          {href ? (
            <Link href={href} variant="body" className="min-w-0">
              {title}
            </Link>
          ) : (
            <Text weight="medium" color="primary" lines={1}>
              {title}
            </Text>
          )}
          {chip ? (
            <Status theme={chip.theme} label={chip.label} variant="chip" />
          ) : null}
          {pendingApproval ? (
            <Status theme="warn" label="Pending approval" variant="chip" />
          ) : null}
        </span>
      }
      status={workflow?.status}
      time={workflow?.created_at}
      {...props}
    >
      <span className="flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1">
        {workflow?.id ? (
          <ID value={workflow.id} label="Copy run ID" truncate />
        ) : null}
        <TriggeredBy workflow={workflow} />
        {workflow?.finished_at ? (
          <Duration
            start={workflow?.started_at ?? workflow?.created_at}
            end={workflow.finished_at}
            variant="caption"
            color="tertiary"
          />
        ) : null}
      </span>
      {badges.length ? (
        <span className="flex min-w-0 flex-wrap items-center gap-1">
          {badges.map((entry) =>
            entry.theme ? (
              <Status
                key={entry.id}
                theme={entry.theme}
                label={entry.label}
                variant="chip"
              />
            ) : (
              <Badge key={entry.id} variant="code">
                {entry.label}
              </Badge>
            )
          )}
        </span>
      ) : null}
    </TimelineItem>
  )
}
