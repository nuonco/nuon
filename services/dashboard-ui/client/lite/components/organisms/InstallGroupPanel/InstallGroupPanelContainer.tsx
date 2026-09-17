import { useState } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import {
  getAppInstalls,
  getAppLabels,
  getBranchInstallGroupRuns,
  toLabelColorMap,
} from '@/lib'
import type { TInstall } from '@/types'
import { latestBranchConfig } from '@/utils/branch-utils'
import { useSurfaces } from '../../../hooks/use-surfaces'
import { useApp } from '../../../providers/app-provider'
import { useAppBranch } from '../../../providers/app-branch-provider'
import { useOrg } from '../../../providers/org-provider'
import { resolveDeploymentPlanStages } from '../../../utils/deployment-plan'
import { InstallGroupPanel } from './InstallGroupPanel'

const INSTALL_PAGE_SIZE = 100
const PANEL_PAGE_SIZE = 20

export const InstallGroupPanelContainer = ({
  groupId,
}: {
  groupId?: string
}) => {
  const { orgId } = useOrg()
  const { appId } = useApp()
  const { branch, branchId, loading: branchLoading } = useAppBranch()
  const { openPanelKey } = useSurfaces()
  const [offset, setOffset] = useState(0)
  const config = branch ? latestBranchConfig(branch) : undefined
  const groups = config?.install_groups ?? []
  const runId = branch?.latest_run?.id

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
      let pageOffset = 0

      for (;;) {
        const page = await getAppInstalls({
          orgId: orgId!,
          appId: appId!,
          app_branch_id: branchId!,
          limit: INSTALL_PAGE_SIZE,
          offset: pageOffset,
        })
        const rows = page?.data ?? []
        installs.push(...rows)
        if (!page?.pagination?.hasNext || rows.length === 0) break

        const nextOffset =
          Number(page.pagination.offset ?? pageOffset) +
          Number(page.pagination.limit ?? INSTALL_PAGE_SIZE)
        if (nextOffset <= pageOffset) break
        pageOffset = nextOffset
      }

      return installs
    },
    enabled: !!orgId && !!appId && !!branchId && groups.length > 0,
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
  const stage = stages.find((item) => item.id === groupId)
  const run = groupRunsQuery.data?.find(
    (item) => item.install_group_id === groupId
  )
  const loading =
    branchLoading ||
    installsQuery.isLoading ||
    (!!runId && groupRunsQuery.isLoading)
  const error =
    installsQuery.error ??
    groupRunsQuery.error ??
    (!loading && !stage ? new Error('Install group unavailable') : undefined)

  return (
    <InstallGroupPanel
      stage={stage}
      run={run}
      labelColors={toLabelColorMap(labelsQuery.data)}
      offset={offset}
      pageSize={PANEL_PAGE_SIZE}
      onOffsetChange={setOffset}
      onInstallSelect={(install) => openPanelKey(`install:${install.id}`)}
      loading={loading}
      error={error}
    />
  )
}
