import { PullRequestScope } from './PullRequestScope'
import {
  configuredPullRequest,
  draftPullRequest,
  firstRunOnly,
  mergedPullRequest,
  needsInputPullRequest,
  previewRuns,
  readyFromDefaultsPullRequest,
  runningRun,
} from './fixtures'

export default {
  title: 'Playground/PullRequest/Scope',
}

/** The whole flow: comment → configure → detail. */
export const Default = () => <PullRequestScope />

/** Branch defaults have no install, so the PR must pick one before it can plan. */
export const FirstVisitNeedsInstall = () => (
  <PullRequestScope
    pullRequest={needsInputPullRequest}
    runs={firstRunOnly}
    initialStep="configure"
  />
)

/** Branch defaults already resolve, so the first run planned without any input. */
export const FirstVisitDefaultsReady = () => (
  <PullRequestScope
    pullRequest={readyFromDefaultsPullRequest}
    runs={[{ ...previewRuns[2], run_number: 1 }]}
    initialStep="detail"
  />
)

/** Every later visit: five runs across three commits, mixed statuses. */
export const ReturnVisitWithRuns = () => (
  <PullRequestScope
    pullRequest={configuredPullRequest}
    runs={previewRuns}
    initialStep="detail"
  />
)

export const RunInProgress = () => (
  <PullRequestScope
    pullRequest={configuredPullRequest}
    runs={[runningRun, ...previewRuns]}
    initialStep="detail"
  />
)

export const ApplyMode = () => (
  <PullRequestScope
    pullRequest={configuredPullRequest}
    runs={previewRuns.filter((run) => run.mode === 'apply')}
    initialStep="detail"
  />
)

/** A draft PR: build and validate only, no install needed. */
export const BuildOnly = () => (
  <PullRequestScope
    pullRequest={draftPullRequest}
    runs={previewRuns.filter((run) => run.mode === 'build-only')}
    initialStep="detail"
  />
)

/** Merged PR: history stays readable, actions are gone. */
export const ClosedMerged = () => (
  <PullRequestScope
    pullRequest={mergedPullRequest}
    runs={previewRuns}
    initialStep="detail"
  />
)

export const Empty = () => (
  <PullRequestScope
    pullRequest={readyFromDefaultsPullRequest}
    runs={[]}
    initialStep="detail"
  />
)

/** Just the comment, for reviewing the GitHub-side copy on its own. */
export const CommentOnly = () => (
  <PullRequestScope
    pullRequest={needsInputPullRequest}
    runs={firstRunOnly}
    initialStep="comment"
  />
)
