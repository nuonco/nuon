import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { previewModeDisplayLabel } from '@/components/branches/shared/preview-mode'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TAppBranchRunPreviewMode } from '@/types'
import type {
  TPlaygroundPullRequest,
  TPlaygroundPullRequestRun,
} from './fixtures'

const TRIGGER_LABEL: Record<TPlaygroundPullRequestRun['trigger'], string> = {
  pull_request: 'Pull request',
  push: 'Push',
  manual: 'Re-run',
}

const MODE_THEME: Record<TAppBranchRunPreviewMode, 'brand' | 'info' | 'neutral'> =
  {
    none: 'neutral',
    apply: 'brand',
    'plan-only': 'info',
    'build-only': 'neutral',
  }

export const ImpactCell = ({ run }: { run: TPlaygroundPullRequestRun }) => {
  if (run.mode === 'build-only') {
    if (run.components_built == null) {
      return (
        <Text variant="subtext" theme="neutral">
          —
        </Text>
      )
    }
    return (
      <Text variant="subtext" theme="neutral">
        {run.components_built} built
      </Text>
    )
  }

  if (
    run.status !== 'success' ||
    run.added == null ||
    run.changed == null ||
    run.removed == null
  ) {
    return (
      <Text variant="subtext" theme="neutral">
        —
      </Text>
    )
  }

  return (
    <span className="flex items-center gap-1 whitespace-nowrap">
      <Badge size="xs" theme="success">
        +{run.added}
      </Badge>
      <Badge size="xs" theme="warn">
        ~{run.changed}
      </Badge>
      <Badge size="xs" theme="error">
        -{run.removed}
      </Badge>
    </span>
  )
}

// A factory, not a component: SurfacesProvider clones whatever it is handed to
// inject isVisible, so the Panel has to be the outermost element.
export const pullRequestRunPanel = (
  pullRequest: TPlaygroundPullRequest,
  run: TPlaygroundPullRequestRun
) => (
  <Panel heading={`Preview run #${run.run_number}`} size="half">
    <div className="flex flex-col gap-6">
      <div className="grid grid-cols-2 gap-4">
        <div className="flex flex-col gap-1">
          <Text variant="subtext" theme="neutral">
            Status
          </Text>
          <Status variant="default" status={run.status} />
        </div>
        <div className="flex flex-col gap-1">
          <Text variant="subtext" theme="neutral">
            Mode
          </Text>
          <Badge size="sm" theme={MODE_THEME[run.mode]}>
            {previewModeDisplayLabel(run.mode)}
          </Badge>
        </div>
        <div className="flex flex-col gap-1">
          <Text variant="subtext" theme="neutral">
            Install
          </Text>
          <Text variant="subtext" family={run.install_name ? 'mono' : 'sans'}>
            {run.install_name ?? 'Not used'}
          </Text>
        </div>
        <div className="flex flex-col gap-1">
          <Text variant="subtext" theme="neutral">
            Duration
          </Text>
          <Duration
            variant="subtext"
            beginTime={run.started_at}
            endTime={run.completed_at}
          />
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <Text variant="subtext" theme="neutral">
          Commit
        </Text>
        <BranchRunCommit
          showStatus={false}
          sha={run.head_sha}
          message={run.commit_message}
          author={run.commit_author}
          createdAt={run.started_at}
          href={
            pullRequest.html_url
              ? `${pullRequest.html_url}/commits/${run.head_sha}`
              : undefined
          }
          isExternal={!!pullRequest.html_url}
        />
      </div>

      {run.error_message ? (
        <div className="flex flex-col gap-2">
          <Text variant="subtext" theme="neutral">
            Error
          </Text>
          <pre className="whitespace-pre-wrap rounded-md border border-red-300 bg-red-50 p-3 font-mono text-xs text-red-800 dark:border-red-500/40 dark:bg-red-950 dark:text-red-500">
            {run.error_message}
          </pre>
        </div>
      ) : null}
    </div>
  </Panel>
)

const byNewest = (a: TPlaygroundPullRequestRun, b: TPlaygroundPullRequestRun) =>
  Date.parse(b.started_at) - Date.parse(a.started_at)

export const groupRunsByPush = (runs: TPlaygroundPullRequestRun[]) => {
  const bySha = new Map<string, TPlaygroundPullRequestRun[]>()
  for (const run of runs) {
    const group = bySha.get(run.head_sha) ?? []
    group.push(run)
    bySha.set(run.head_sha, group)
  }
  return [...bySha.values()]
    .map((group) => [...group].sort(byNewest))
    .sort((a, b) => byNewest(a[0], b[0]))
}

const latestPlannedChange = (runs: TPlaygroundPullRequestRun[]) =>
  [...runs]
    .filter((run) => run.mode !== 'build-only' && run.status === 'success')
    .sort(byNewest)[0]

const countLabel = (count: number, singular: string, plural: string) =>
  `${count} ${count === 1 ? singular : plural}`

export interface IPullRequestActivity {
  pullRequest: TPlaygroundPullRequest
  runs: TPlaygroundPullRequestRun[]
  isLoading?: boolean
}

const SKELETON_PUSHES = [0, 1]
const SKELETON_RUNS_PER_PUSH = 2

export const PullRequestActivity = ({
  pullRequest,
  runs,
  isLoading = false,
}: IPullRequestActivity) => {
  const { addPanel } = useSurfaces()
  const pushes = groupRunsByPush(runs)
  const latest = latestPlannedChange(runs)

  return (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-3">
        <LabeledValue label="Pushes" loading={isLoading}>
          <Text variant="subtext">{pushes.length}</Text>
        </LabeledValue>
        <LabeledValue label="Runs" loading={isLoading}>
          <Text variant="subtext">{runs.length}</Text>
        </LabeledValue>
        <LabeledValue label="Latest change" loading={isLoading}>
          {latest ? (
            <ImpactCell run={latest} />
          ) : (
            <Text variant="subtext" theme="neutral">
              No planned changes yet
            </Text>
          )}
        </LabeledValue>
      </div>

      {isLoading ? (
        SKELETON_PUSHES.map((push) => (
          <Card key={push}>
            <div className="flex items-start justify-between gap-3">
              <Text variant="body" loading loadingWidth={28} />
              <Badge size="sm" loading loadingWidth={10} />
            </div>
            <div className="flex flex-col divide-y border-t">
              {Array.from({ length: SKELETON_RUNS_PER_PUSH }, (_, run) => (
                <div
                  key={run}
                  className="flex w-full flex-wrap items-center gap-x-3 gap-y-1 px-1 py-2"
                >
                  <Text variant="subtext" loading loadingWidth={6} />
                  <Badge size="xs" loading loadingWidth={12} />
                  <Badge size="xs" loading loadingWidth={10} />
                  <Text variant="subtext" loading loadingWidth={14} />
                  <span className="ml-auto">
                    <Text variant="subtext" loading loadingWidth={8} />
                  </span>
                </div>
              ))}
            </div>
          </Card>
        ))
      ) : pushes.length === 0 ? (
        <EmptyState
          variant="history"
          emptyTitle="No preview runs yet"
          emptyMessage="Activity shows up here when this pull request is opened or pushed to."
        />
      ) : (
        pushes.map((pushRuns) => {
          const commit = pushRuns[0]
          return (
            <Card key={commit.head_sha}>
              <div className="flex items-start justify-between gap-3">
                <BranchRunCommit
                  showStatus={false}
                  className="min-w-0"
                  sha={commit.head_sha}
                  message={commit.commit_message}
                  author={commit.commit_author}
                  createdAt={commit.started_at}
                  href={
                    pullRequest.html_url
                      ? `${pullRequest.html_url}/commits/${commit.head_sha}`
                      : undefined
                  }
                  isExternal={!!pullRequest.html_url}
                />
                <Badge size="sm" theme="neutral">
                  {countLabel(pushRuns.length, 'run', 'runs')}
                </Badge>
              </div>
              <div className="flex flex-col divide-y border-t">
                {pushRuns.map((run) => (
                  <button
                    key={run.id}
                    type="button"
                    className="flex w-full cursor-pointer flex-wrap items-center gap-x-3 gap-y-1 px-1 py-2 text-left hover:bg-cool-grey-500/8 focus:outline-none focus:shadow-[0_0_0_1px_var(--background),0_0_0_3px_var(--focus-ring)]"
                    onClick={() =>
                      addPanel(pullRequestRunPanel(pullRequest, run))
                    }
                  >
                    <span className="flex items-center gap-1.5">
                      <Status variant="default" status={run.status} />
                      <Text variant="subtext" family="mono">
                        #{run.run_number}
                      </Text>
                    </span>
                    <Badge size="xs" theme="neutral">
                      {TRIGGER_LABEL[run.trigger]}
                    </Badge>
                    <Badge size="xs">{previewModeDisplayLabel(run.mode)}</Badge>
                    <Text variant="subtext" family="mono" theme="neutral">
                      {run.install_name ?? 'No install'}
                    </Text>
                    <ImpactCell run={run} />
                    <span className="ml-auto flex items-center gap-3">
                      <Time
                        variant="subtext"
                        theme="neutral"
                        time={run.started_at}
                        format="relative"
                      />
                      {run.completed_at ? (
                        <Duration
                          variant="subtext"
                          theme="neutral"
                          beginTime={run.started_at}
                          endTime={run.completed_at}
                        />
                      ) : (
                        <Text variant="subtext" theme="neutral">
                          in progress
                        </Text>
                      )}
                    </span>
                  </button>
                ))}
              </div>
            </Card>
          )
        })
      )}
    </div>
  )
}
