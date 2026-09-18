import { AppBranchRunCard } from '@/components/branches/AppBranchRunCard'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RunSummary } from '@/components/runs/RunSummary'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSandboxBuild } from '@/hooks/use-sandbox-build'

export const SandboxBuildSummaryTab = () => {
  const { build } = useSandboxBuild()
  const { app } = useApp()
  const { org } = useOrg()

  const jobs = build?.runner_job ? [build.runner_job] : []

  return (
    <>
      <PageTitle segments={['Sandbox build', app?.name]} />
      <AppBranchRunCard
        appId={app?.id}
        orgId={org?.id}
        buildStatus={build?.status_v2?.status}
        sourceCommit={build?.vcs_connection_commit}
        run={build?.app_branch_run}
      />
      <RunSummary
        isLoading={!build}
        status={build?.status_v2}
        statusDescription={build?.status_description}
        showTiming={false}
        jobs={jobs}
        jobHref={(job) => `/${org?.id}/runner/jobs/${job?.id}`}
      />
    </>
  )
}
