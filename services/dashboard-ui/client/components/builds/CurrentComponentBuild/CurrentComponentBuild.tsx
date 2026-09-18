import { AppBranchRunCard } from '@/components/branches/AppBranchRunCard'
import { BuildImageSource } from '@/components/builds/BuildImageSource'
import { CompositeError } from '@/components/common/CompositeError'
import { Link } from '@/components/common/Link'
import { LogsPanel } from '@/components/log-stream/LogsPanel'
import { RunFailureBanner } from '@/components/runs/RunFailureBanner'
import type { TBuild } from '@/types'
import { isImageBuild } from '@/utils/image-ref'

export interface ICurrentComponentBuild {
  appId?: string
  orgId?: string
  build: TBuild
  buildHref?: string
}

export const CurrentComponentBuild = ({
  appId,
  orgId,
  build,
  buildHref,
}: ICurrentComponentBuild) => {
  const status = build.status_v2?.status
    ? build.status_v2
    : { status: build.status }

  return (
    <div className="flex flex-col gap-4">
      <AppBranchRunCard
        appId={appId}
        orgId={orgId}
        buildStatus={status.status}
        sourceCommit={build.vcs_connection_commit}
        sourceHref={buildHref}
        run={build.app_branch_run}
      />

      {build.composite_error ? (
        <CompositeError error={build.composite_error} />
      ) : null}

      <RunFailureBanner
        jobs={build.runner_job ? [build.runner_job] : []}
        status={status}
        statusDescription={build.status_description}
      />

      {isImageBuild(build) ? <BuildImageSource build={build} /> : null}

      <LogsPanel logStream={build.log_stream} />

      {buildHref ? <Link href={buildHref}>View build</Link> : null}
    </div>
  )
}
