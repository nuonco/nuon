import { AppBranchRunCard } from '@/components/branches/AppBranchRunCard'
import { CompositeError } from '@/components/common/CompositeError'
import { Link } from '@/components/common/Link'
import { LogsPanel } from '@/components/log-stream/LogsPanel'
import { RunSummary } from '@/components/runs/RunSummary'
import type { TAppSandboxBuild } from '@/types'

export interface ICurrentSandboxBuild {
  appId?: string
  orgId?: string
  build: TAppSandboxBuild
  buildHref?: string
  sourceRepo?: string
}

export const CurrentSandboxBuild = ({
  appId,
  orgId,
  build,
  buildHref,
  sourceRepo,
}: ICurrentSandboxBuild) => {
  const jobs = build.runner_job ? [build.runner_job] : []

  return (
    <div className="flex flex-col gap-4">
      <AppBranchRunCard
        appId={appId}
        orgId={orgId}
        buildStatus={build.status_v2?.status}
        sourceCommit={build.vcs_connection_commit}
        sourceHref={buildHref}
        sourceRepo={sourceRepo}
        run={build.app_branch_run}
      />

      {build.composite_error ? (
        <CompositeError error={build.composite_error} />
      ) : null}

      <RunSummary
        status={build.status_v2}
        statusDescription={build.status_description}
        showTiming={false}
        jobs={jobs}
        jobHref={(job) =>
          orgId ? `/${orgId}/runner/jobs/${job?.id}` : undefined
        }
      />

      <LogsPanel logStream={build.log_stream} />

      {buildHref ? <Link href={buildHref}>View build</Link> : null}
    </div>
  )
}
