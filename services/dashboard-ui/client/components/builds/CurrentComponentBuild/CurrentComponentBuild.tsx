import { AppBranchRunCard } from '@/components/branches/AppBranchRunCard'
import { BuildImageSource } from '@/components/builds/BuildImageSource'
import { CompositeError } from '@/components/common/CompositeError'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { LogsPanel } from '@/components/log-stream/LogsPanel'
import { RunSummary } from '@/components/runs/RunSummary'
import type { TBuild } from '@/types'
import { isImageBuild } from '@/utils/image-ref'

export interface ICurrentComponentBuild {
  appId?: string
  orgId?: string
  build: TBuild
  buildHref?: string
  sourceRepo?: string
}

export const CurrentComponentBuild = ({
  appId,
  orgId,
  build,
  buildHref,
  sourceRepo,
}: ICurrentComponentBuild) => {
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
        timings={[
          { label: 'Created', time: build.created_at },
          ...(build.resolved_at
            ? [{ label: 'Resolved', time: build.resolved_at }]
            : []),
          { label: 'Updated', time: build.updated_at },
        ]}
        duration={{ beginTime: build.created_at, endTime: build.updated_at }}
        jobs={jobs}
        jobHref={(job) =>
          orgId ? `/${orgId}/runner/jobs/${job?.id}` : undefined
        }
        triggeredBy={
          build.created_by?.email ? (
            <Text variant="subtext">{build.created_by.email}</Text>
          ) : build.created_by_id ? (
            <ID>{build.created_by_id}</ID>
          ) : null
        }
      >
        {isImageBuild(build) ? <BuildImageSource build={build} /> : null}
      </RunSummary>

      <LogsPanel logStream={build.log_stream} />

      {buildHref ? <Link href={buildHref}>View build</Link> : null}
    </div>
  )
}
