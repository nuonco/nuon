import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Loading } from '@/components/common/Loading'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { previewModeDisplayLabel } from '@/components/branches/shared/preview-mode'
import { useSurfaces } from '@/hooks/use-surfaces'
import { ImpactCell, pullRequestRunPanel } from './PullRequestRunsTable'
import type {
  TPlaygroundPullRequest,
  TPlaygroundPullRequestRun,
} from './fixtures'

const TRIGGER_LABEL: Record<TPlaygroundPullRequestRun['trigger'], string> = {
  pull_request: 'Pull request',
  push: 'Push',
  manual: 'Re-run',
}

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
        <LabeledValue label="Pushes">
          <Text variant="subtext">{pushes.length}</Text>
        </LabeledValue>
        <LabeledValue label="Runs">
          <Text variant="subtext">{runs.length}</Text>
        </LabeledValue>
        <LabeledValue label="Latest change">
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
        <div className="flex justify-center py-8">
          <Loading />
        </div>
      ) : pushes.length === 0 ? (
        <EmptyState
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
                  href={`${pullRequest.html_url}/commits/${commit.head_sha}`}
                  isExternal
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
