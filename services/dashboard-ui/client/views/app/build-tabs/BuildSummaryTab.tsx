import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { BuildImageSource } from '@/components/builds/BuildImageSource'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RunSummary } from '@/components/runs/RunSummary'
import { useApp } from '@/hooks/use-app'
import { useBuild } from '@/hooks/use-build'
import { useOrg } from '@/hooks/use-org'
import { isImageBuild } from '@/utils/image-ref'

export const BuildSummaryTab = () => {
  const { build } = useBuild()
  const { app } = useApp()
  const { org } = useOrg()

  const jobs = build?.runner_job ? [build.runner_job] : []

  return (
    <>
      <PageTitle segments={['Build', app?.name]} />
      <RunSummary
        isLoading={!build}
        status={build?.status_v2}
        statusDescription={build?.status_description}
        timings={[
          { label: 'Created', time: build?.created_at },
          ...(build?.resolved_at
            ? [{ label: 'Resolved', time: build.resolved_at }]
            : []),
          { label: 'Updated', time: build?.updated_at },
        ]}
        duration={{ beginTime: build?.created_at, endTime: build?.updated_at }}
        jobs={jobs}
        jobHref={(job) => `/${org?.id}/runner/jobs/${job?.id}`}
        triggeredBy={
          build?.created_by?.email ? (
            <Text variant="subtext">{build.created_by.email}</Text>
          ) : build?.created_by_id ? (
            <ID>{build.created_by_id}</ID>
          ) : null
        }
      >
        {build && isImageBuild(build) ? <BuildImageSource build={build} /> : null}
      </RunSummary>
    </>
  )
}
