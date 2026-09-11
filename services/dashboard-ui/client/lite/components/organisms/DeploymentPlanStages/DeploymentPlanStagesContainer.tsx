import { keepPreviousData, useQuery } from '@tanstack/react-query'
import {
  getAppInstalls,
  getAppLabels,
  getBranchInstallGroupRuns,
  toLabelColorMap,
} from '@/lib'
import type { TInstall } from '@/types'
import { latestBranchConfig } from '@/utils/branch-utils'
import { usePanelHref, useSurfaces } from '../../../hooks/use-surfaces'
import { useApp } from '../../../providers/app-provider'
import { useAppBranch } from '../../../providers/app-branch-provider'
import { useOrg } from '../../../providers/org-provider'
import {
  resolveDeploymentPlanStages,
  type IDeploymentPlanStage,
} from '../../../utils/deployment-plan'
import { InstallGroupCard } from '../InstallGroupCard'
import { DeploymentPlanStages } from './DeploymentPlanStages'

const INSTALL_PAGE_SIZE = 100
const RUN_POLL_INTERVAL = 20_000
const TERMINAL_RUN_STATUSES = new Set([
  'success',
  'error',
  'cancelled',
  'not-attempted',
])

const LinkableInstallGroupCard = ({
  stage,
  labelColors,
}: {
  stage: IDeploymentPlanStage
  labelColors: Record<string, string>
}) => {
  const groupHref = usePanelHref(`group:${stage.id}`)
  const { openPanelKey } = useSurfaces()

  return (
    <InstallGroupCard
      stage={stage}
      labelColors={labelColors}
      groupHref={groupHref}
      onInstallSelect={(install) => openPanelKey(`install:${install.id}`)}
    />
  )
}

export const DeploymentPlanStagesContainer = () => {
  const { orgId } = useOrg()
  const { appId } = useApp()
  const {
    branch,
    branchId,
    loading: branchLoading,
    error: branchError,
  } = useAppBranch()
  const config = branch ? latestBranchConfig(branch) : undefined
  const groups = config?.install_groups ?? []
  const runId = branch?.latest_run?.id
  const runStatus = branch?.latest_run?.status

  const installsQuery = useQuery({
    queryKey: [
      'app-installs',
      orgId,
      appId,
      branchId,
      'deployment-plan',
    ],
    queryFn: async () => {
      const installs: TInstall[] = []
      let offset = 0

      for (;;) {
        const page = await getAppInstalls({
          orgId: orgId!,
          appId: appId!,
          app_branch_id: branchId!,
          limit: INSTALL_PAGE_SIZE,
          offset,
        })
        const rows = page?.data ?? []
        installs.push(...rows)
        if (!page?.pagination?.hasNext || rows.length === 0) break

        const nextOffset =
          Number(page.pagination.offset ?? offset) +
          Number(page.pagination.limit ?? INSTALL_PAGE_SIZE)
        if (nextOffset <= offset) break
        offset = nextOffset
      }

      return installs
    },
    enabled:
      !!orgId && !!appId && !!branchId && groups.length > 0,
    placeholderData: keepPreviousData,
  })

  const groupRunsQuery = useQuery({
    queryKey: [
      'branch-install-group-runs',
      orgId,
      appId,
      branchId,
      runId,
    ],
    queryFn: () =>
      getBranchInstallGroupRuns({
        orgId: orgId!,
        appId: appId!,
        branchId: branchId!,
        runId: runId!,
      }),
    enabled: !!orgId && !!appId && !!branchId && !!runId,
    placeholderData: keepPreviousData,
    refetchInterval:
      runId && !TERMINAL_RUN_STATUSES.has(runStatus ?? '')
        ? RUN_POLL_INTERVAL
        : false,
  })

  const labelsQuery = useQuery({
    queryKey: ['app-labels', orgId, appId],
    queryFn: () => getAppLabels({ orgId: orgId!, appId: appId! }),
    enabled: !!orgId && !!appId,
    staleTime: 60_000,
  })

  const stages = resolveDeploymentPlanStages({
    groups,
    installs: installsQuery.data,
    groupRuns: groupRunsQuery.data,
  })
  const loading =
    branchLoading ||
    (groups.length > 0 &&
      (installsQuery.isLoading || (!!runId && groupRunsQuery.isLoading)))
  const error =
    branchError ?? installsQuery.error ?? groupRunsQuery.error
  const labelColors = toLabelColorMap(labelsQuery.data)

  return (
    <DeploymentPlanStages
      branch={branch}
      stages={stages}
      labelColors={labelColors}
      renderStageCard={(stage) => (
        <LinkableInstallGroupCard
          key={stage.id}
          stage={stage}
          labelColors={labelColors}
        />
      )}
      loading={loading}
      error={error}
    />
  )
}
