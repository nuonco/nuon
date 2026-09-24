import { useState } from 'react'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import type { TAppBranchPreviewConfig } from '@/types'
import { PullRequestComment } from './PullRequestComment'
import { PullRequestConfigure } from './PullRequestConfigure'
import { PullRequestDetail } from './PullRequestDetail'
import {
  firstRunOnly,
  needsInputPullRequest,
  pullRequestPath,
  type TPlaygroundPullRequest,
  type TPlaygroundPullRequestRun,
} from './fixtures'

export type TPullRequestStep = 'comment' | 'configure' | 'detail'

/**
 * Walks the whole PR scope flow in one story:
 *
 *   1. the sticky GitHub comment posted when the PR opens
 *   2. the configure screen it links into, where you pick install + mode
 *   3. the detail screen every later visit lands on, showing run + commit
 *      history for the PR
 *
 * Saving on step 2 promotes the PR to config_status 'configured' and appends a
 * re-run, which is what the real PATCH .../preview-config + POST .../runs pair
 * would do.
 */
export interface IPullRequestScope {
  pullRequest?: TPlaygroundPullRequest
  runs?: TPlaygroundPullRequestRun[]
  initialStep?: TPullRequestStep
}

export const PullRequestScope = ({
  pullRequest: initialPullRequest = needsInputPullRequest,
  runs: initialRuns = firstRunOnly,
  initialStep = 'comment',
}: IPullRequestScope) => {
  const [step, setStep] = useState<TPullRequestStep>(initialStep)
  const [pullRequest, setPullRequest] = useState(initialPullRequest)
  const [runs, setRuns] = useState(initialRuns)

  const appendRun = (resolved: TAppBranchPreviewConfig) => {
    const previous = runs[0]
    const next: TPlaygroundPullRequestRun = {
      id: `abrun-rerun-${runs.length + 1}`,
      run_number: (previous?.run_number ?? 0) + 1,
      status: 'running',
      trigger: 'manual',
      mode: resolved.mode ?? 'plan-only',
      install_id: resolved.install_id,
      install_name: resolved.install_name,
      head_sha: pullRequest.head_sha,
      commit_message: previous?.commit_message ?? 'Latest commit on this PR',
      commit_author: pullRequest.author_login,
      started_at: new Date().toISOString(),
      added: 0,
      changed: 0,
      removed: 0,
      components_built: 0,
    }
    setRuns([next, ...runs])
  }

  const onSave = (override: TAppBranchPreviewConfig) => {
    const resolved: TAppBranchPreviewConfig = {
      ...pullRequest.branch_preview_config,
      ...override,
    }
    setPullRequest({
      ...pullRequest,
      config_status: 'configured',
      preview_override: override,
      resolved_preview_config: resolved,
    })
    appendRun(resolved)
    setStep('detail')
  }

  if (step === 'comment') {
    return (
      <PageSection className="items-start gap-3">
        <Text variant="subtext" theme="neutral">
          Step 1 — the sticky comment Nuon posts when the pull request opens.
          Both buttons land on{' '}
          <Text variant="subtext" theme="neutral" family="mono">
            {pullRequestPath(pullRequest)}
          </Text>
          .
        </Text>
        <PullRequestComment
          pullRequest={pullRequest}
          run={runs[0]}
          runCount={runs.length}
          onConfigure={() => setStep('configure')}
          onViewRuns={() => setStep('detail')}
        />
      </PageSection>
    )
  }

  if (step === 'configure') {
    return (
      <PullRequestConfigure
        pullRequest={pullRequest}
        onSave={onSave}
        onCancel={() => setStep('detail')}
      />
    )
  }

  return (
    <PullRequestDetail
      pullRequest={pullRequest}
      runs={runs}
      onEditConfig={() => setStep('configure')}
      onRerun={() => appendRun(pullRequest.resolved_preview_config)}
    />
  )
}
