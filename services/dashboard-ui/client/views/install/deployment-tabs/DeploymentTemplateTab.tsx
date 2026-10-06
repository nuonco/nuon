import { useMemo } from 'react'
import { useOutletContext } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import {
  changedBuildRows,
  type TBuildMeta,
} from '@/components/branches/BranchOverview/changed-builds'
import { BuildChangeCards } from '@/components/branches/BranchOverview/RunBuildsPanel'
import { BranchRunChangesSummary } from '@/components/branches/BranchRunChangesSummary'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstallPage } from '@/hooks/use-install-path'
import { useWorkflow } from '@/hooks/use-workflow'
import { getBranchRunBuilds, getBranchWorkflowRun } from '@/lib'
import { AppProvider } from '@/providers/app-provider'
import type { TInstallDeploymentRecord } from '@/types'
import { humanize } from '@/utils/string-utils'

const isBuildStep = (name?: string) =>
  !!name && /build/i.test(name) && !/(?:fetch|sync) app config/i.test(name)

const NO_BUILDS: TBuildMeta[] = []

export const DeploymentTemplateTab = () => {
  const { org, install } = useInstallPage()
  const { workflow } = useWorkflow()
  const { deployment, deploymentFetched } = useOutletContext<{
    deployment?: TInstallDeploymentRecord
    deploymentFetched: boolean
  }>()
  const title =
    deployment?.title ||
    workflow?.name ||
    humanize(workflow?.type) ||
    'Deployment'
  const branch = deployment?.app_branch
  const ready = !!branch?.id && !!branch.run_id && !!install?.app_id && !!org?.id

  const { data: branchRun } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'branch-workflow-run',
      org?.id,
      install?.app_id,
      branch?.id,
      branch?.run_id,
    ],
    queryFn: () =>
      getBranchWorkflowRun({
        orgId: org!.id,
        appId: install!.app_id!,
        branchId: branch!.id,
        runId: branch!.run_id!,
      }),
    enabled: ready,
  })

  const { data: runBuilds } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'branch-run-builds',
      org?.id,
      install?.app_id,
      branch?.id,
      branch?.run_id,
    ],
    queryFn: () =>
      getBranchRunBuilds({
        orgId: org!.id,
        appId: install!.app_id!,
        branchId: branch!.id,
        runId: branch!.run_id!,
      }),
    enabled: ready,
  })

  const buildStep = branchRun?.steps?.find((step) => isBuildStep(step.name))
  const buildMetadata = (buildStep?.status?.metadata ?? {}) as {
    builds?: TBuildMeta[]
    sandbox_build_id?: string
  }
  const metaBuilds = buildMetadata.builds ?? NO_BUILDS
  const changedBuilds = useMemo(
    () =>
      changedBuildRows({
        metaBuilds,
        runBuilds: (runBuilds ?? []).map((build) => ({
          id: build.id,
          component_id: build.component_id,
          component_name: build.component_name,
          status: build.status_v2?.status || build.status,
        })),
        orgId: org?.id,
        appId: install?.app_id,
        sandboxBuildId: buildMetadata.sandbox_build_id,
      }),
    [metaBuilds, runBuilds, org?.id, install?.app_id, buildMetadata.sandbox_build_id]
  )

  return (
    <>
      <PageTitle segments={[title, install?.name]} />
      {!deploymentFetched ? (
        <Text loading loadingWidth={32} />
      ) : ready ? (
        <AppProvider appId={install.app_id!}>
          <div className="flex flex-col gap-4">
            <BranchRunChangesSummary
              branchId={branch.id}
              appBranchRunId={branch.run_id!}
              builds={metaBuilds}
              title="Template updates"
            />
            <BuildChangeCards rows={changedBuilds} />
          </div>
        </AppProvider>
      ) : (
        <EmptyState
          emptyTitle="No template updates"
          emptyMessage="This deployment has no app config changes."
        />
      )}
    </>
  )
}
