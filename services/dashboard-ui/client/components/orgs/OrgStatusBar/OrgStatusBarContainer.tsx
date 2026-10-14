import { useMemo } from 'react'
import { keepPreviousData, useQueries, useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { type TContextTooltipItem } from '@/components/common/ContextTooltip'
import { Status } from '@/components/common/Status'
import { useActiveWorkflows } from '@/hooks/use-active-workflows'
import { useConfig } from '@/hooks/use-config'
import { useOrg } from '@/hooks/use-org'
import { useInstallNested } from '@/hooks/use-install-path'
import { useWorkflowApprovals } from '@/hooks/use-workflow-approvals'
import {
  getApp,
  getAppBranch,
  getAppConfigs,
  getInstall,
  getInstallStack,
} from '@/lib'
import { humanize } from '@/utils/string-utils'
import {
  getApprovalHref,
  getWorkflowHref,
  getWorkflowStepTitle,
} from '@/utils/workflow-utils'
import { OrgStatusBar } from './OrgStatusBar'

export const OrgStatusBarContainer = () => {
  const nested = useInstallNested()
  const { org } = useOrg()
  const { byocName, byocColor, byocTextColor } = useConfig()
  const { approvals } = useWorkflowApprovals()
  const { activeWorkflows } = useActiveWorkflows()
  const { appId, branchId, installId } = useParams()

  const { data: appData } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app', org.id, appId],
    queryFn: () => getApp({ orgId: org.id, appId: appId! }),
    enabled: !!appId,
  })

  const { data: appConfigs } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-configs', org.id, appId],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: appId!, limit: 1 }),
    enabled: !!appId,
    refetchInterval: 30_000,
  })

  const { data: installData } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install', org.id, installId],
    queryFn: () => getInstall({ orgId: org.id, installId: installId! }),
    enabled: !!installId,
  })
  const install = installId ? installData : undefined
  const resolvedBranchId = branchId ?? install?.app_branch_id

  const { data: branchData } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-branch', org.id, appId, resolvedBranchId],
    queryFn: () => getAppBranch({ orgId: org.id, appId: appId!, branchId: resolvedBranchId! }),
    enabled: !!appId && !!resolvedBranchId,
  })

  const { data: stackData } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-stack', org.id, installId],
    queryFn: () => getInstallStack({ installId: installId!, orgId: org.id }),
    enabled: !!installId,
    refetchInterval: 30_000,
  })

  const app = appId ? appData : undefined
  const latestConfig = appId ? appConfigs?.[0] : undefined
  const branch = appId && resolvedBranchId ? branchData : undefined
  const stack = installId ? stackData : undefined

  const installOwnerIds = useMemo(() => {
    if (!nested) return []
    const ids = new Set<string>()
    for (const workflow of activeWorkflows) {
      if (workflow.owner_type === 'installs' && workflow.owner_id) {
        ids.add(workflow.owner_id)
      }
    }
    for (const approval of approvals) {
      if (approval.app_id && approval.app_branch_id) continue
      const step = approval.workflow_step
      const workflow = activeWorkflows.find(
        (item) => item.id === step?.install_workflow_id
      )
      const ownerId =
        workflow?.owner_type === 'installs'
          ? workflow.owner_id
          : step?.owner_type === 'installs'
            ? step.owner_id
            : undefined
      if (ownerId) ids.add(ownerId)
    }
    return [...ids].sort()
  }, [activeWorkflows, approvals, nested])

  const installQueries = useQueries({
    queries: installOwnerIds.map((id) => ({
      placeholderData: keepPreviousData,
      queryKey: ['install', org.id, id],
      queryFn: () => getInstall({ orgId: org.id, installId: id }),
    })),
  })

  const appIdByInstall = new Map<string, string>()
  const branchIdByInstall = new Map<string, string>()
  if (install?.id && install.app_id) {
    appIdByInstall.set(install.id, install.app_id)
    if (install.app_branch_id) branchIdByInstall.set(install.id, install.app_branch_id)
  }
  installQueries.forEach((query, index) => {
    const id = installOwnerIds[index]
    const resolvedAppId = query.data?.app_id
    if (!id || !resolvedAppId) return
    appIdByInstall.set(id, resolvedAppId)
    if (query.data?.app_branch_id) branchIdByInstall.set(id, query.data.app_branch_id)
  })

  const installTarget = (installId?: string) => {
    if (!installId) return undefined
    const app = appIdByInstall.get(installId)
    const branch = branchIdByInstall.get(installId)
    if (!app) return undefined
    return { appId: app, branchId: branch }
  }

  const workflowItems: TContextTooltipItem[] = activeWorkflows.map((workflow) => {
    const target =
      workflow.owner_type === 'installs' ? installTarget(workflow.owner_id) : undefined
    return {
      id: workflow.id ?? '',
      title: workflow.name || humanize(workflow.type),
      subtitle: workflow.metadata?.owner_name || workflow.status?.status || undefined,
      href: getWorkflowHref(org.id, workflow, {
        nested,
        appId: target?.appId,
        branchId: target?.branchId,
      }),
      leftContent: (
        <Status
          status={workflow.status?.status ?? ''}
          isWithoutText
          variant="timeline"
          iconSize={16}
        />
      ),
    }
  })

  const ownerNames = new Map(
    activeWorkflows
      .filter((w) => w.owner_id && w.metadata?.owner_name)
      .map((w) => [w.owner_id!, w.metadata!.owner_name!])
  )

  const approvalItems: TContextTooltipItem[] = approvals.map((approval) => {
    const step = approval.workflow_step
    const workflow = activeWorkflows.find(
      (item) => item.id === step?.install_workflow_id
    )
    const installOwnerId =
      workflow?.owner_type === 'installs'
        ? workflow.owner_id
        : step?.owner_type === 'installs'
          ? step.owner_id
          : undefined
    const target = installTarget(installOwnerId)
    const href = getApprovalHref(org.id, approval, workflow, {
      nested,
      appId: target?.appId,
      branchId: target?.branchId,
    })
    const opensBranch =
      workflow?.owner_type === 'app_branches' ||
      !!workflow?.metadata?.app_branch_run_id ||
      !!(approval.app_id && approval.app_branch_id)
    const stepHref =
      href && step?.id && !opensBranch
        ? `${href}?panel=${encodeURIComponent(step.id)}`
        : href
    const installName = step?.owner_id ? ownerNames.get(step.owner_id) : undefined
    return {
      id: approval.id ?? '',
      title: step?.name ? getWorkflowStepTitle(step) : 'Approval required',
      subtitle: installName || approval.type || undefined,
      href: stepHref,
    }
  })

  return (
    <OrgStatusBar
      org={org}
      app={app}
      branch={branch}
      latestConfig={latestConfig}
      install={install}
      stack={stack}
      approvals={approvals}
      activeWorkflows={activeWorkflows}
      approvalItems={approvalItems}
      workflowItems={workflowItems}
      byocName={byocName}
      byocColor={byocColor}
      byocTextColor={byocTextColor}
    />
  )
}
