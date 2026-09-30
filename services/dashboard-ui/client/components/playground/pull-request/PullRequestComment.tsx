import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { previewModeDisplayLabel } from '@/components/branches/shared/preview-mode'
import type { TPlaygroundPullRequest, TPlaygroundPullRequestRun } from './fixtures'

type TPhase = {
  label: string
  status: string
  detail: string
}

const phasesForRun = (
  pullRequest: TPlaygroundPullRequest,
  run?: TPlaygroundPullRequestRun
): TPhase[] => {
  const mode = run?.mode ?? pullRequest.resolved_preview_config.mode
  const failed = run?.status === 'failed'
  const running = run?.status === 'running'

  const phases: TPhase[] = [
    {
      label: 'Config',
      status: running ? 'running' : 'success',
      detail: running ? 'Parsing nuon.toml' : 'nuon.toml is valid',
    },
    {
      label: 'Builds',
      status: running ? 'pending' : 'success',
      detail: `${run?.components_built ?? 3} components built`,
    },
  ]

  if (mode === 'build-only') return phases

  if (pullRequest.config_status === 'needs-input') {
    phases.push({
      label: 'Install',
      status: 'skipped',
      detail: 'No install selected',
    })
    return phases
  }

  phases.push({
    label: 'Install',
    status: failed ? 'failed' : running ? 'pending' : 'success',
    detail: failed
      ? (run?.error_message ?? 'Apply failed')
      : `${previewModeDisplayLabel(mode ?? 'plan-only')} against ${
          pullRequest.resolved_preview_config.install_name ?? 'the target install'
        }`,
  })

  return phases
}

export interface IPullRequestComment {
  pullRequest: TPlaygroundPullRequest
  run?: TPlaygroundPullRequestRun
  runCount?: number
  onConfigure: () => void
  onViewRuns: () => void
}

/**
 * A mock of the sticky GitHub PR comment, so the story can start where the user
 * starts. Not a real component — it stands in for the markdown that
 * activities/pr_comment_body.go renders.
 */
export const PullRequestComment = ({
  pullRequest,
  run,
  runCount = 1,
  onConfigure,
  onViewRuns,
}: IPullRequestComment) => {
  const needsInput = pullRequest.config_status === 'needs-input'
  const phases = phasesForRun(pullRequest, run)
  const mode = run?.mode ?? pullRequest.resolved_preview_config.mode

  return (
    <Card className="max-w-3xl gap-4 bg-white dark:bg-dark-grey-900">
      <div className="flex items-center gap-2">
        <Icon variant="GithubLogoIcon" size={16} theme="neutral" />
        <Text variant="subtext" theme="neutral">
          <Text variant="subtext" weight="strong">
            nuon
          </Text>{' '}
          commented on {pullRequest.repo_owner}/{pullRequest.repo_name}#
          {pullRequest.number}
        </Text>
      </div>

      <div className="flex flex-col gap-4 border rounded-md p-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <Text variant="base" weight="strong">
            👋 Nuon Preview — {pullRequest.repo_name}
            {mode ? ` (${previewModeDisplayLabel(mode)})` : null}
          </Text>
          {run ? (
            <Status variant="badge" status={run.status} />
          ) : (
            <Status variant="badge" status="pending" />
          )}
        </div>

        {needsInput ? (
          <div className="flex flex-col gap-2 rounded-md border border-orange-400 bg-orange-50 p-3 dark:border-orange-500/40 dark:bg-orange-950">
            <Text variant="body" weight="strong">
              Action required — choose an install
            </Text>
            <Text variant="subtext" theme="neutral">
              This branch has no default preview install, so this run only built
              and validated your config. Pick an install to plan or apply on
              every push to this PR.
            </Text>
          </div>
        ) : null}

        <div className="flex flex-col divide-y">
          {phases.map((phase) => (
            <div
              key={phase.label}
              className="flex items-center justify-between gap-4 py-2 first:pt-0 last:pb-0"
            >
              <span className="flex items-center gap-2 shrink-0">
                <Status variant="default" status={phase.status} />
                <Text variant="subtext" weight="strong">
                  {phase.label}
                </Text>
              </span>
              <Text
                variant="subtext"
                theme="neutral"
                className="truncate text-right"
              >
                {phase.detail}
              </Text>
            </div>
          ))}
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <Text variant="subtext" theme="neutral">
            Commit
          </Text>
          <Badge size="sm" variant="code" theme="neutral">
            {pullRequest.head_sha.slice(0, 7)}
          </Badge>
          <Text variant="subtext" theme="neutral">
            · {runCount} preview {runCount === 1 ? 'run' : 'runs'} on this PR
          </Text>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <Button
            variant={needsInput ? 'primary' : 'secondary'}
            size="md"
            onClick={onConfigure}
          >
            {needsInput ? 'Choose an install →' : 'Configure this preview →'}
          </Button>
          <Button variant="ghost" size="md" onClick={onViewRuns}>
            View preview runs →
          </Button>
        </div>
      </div>
    </Card>
  )
}
