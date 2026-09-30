import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { type TContextTooltipItem } from '@/components/common/ContextTooltip'
import { Status } from '@/components/common/Status'
import { useActiveWorkflows } from '@/hooks/use-active-workflows'
import { useConfig } from '@/hooks/use-config'
import { useOrg } from '@/hooks/use-org'
import { useWorkflowApprovals } from '@/hooks/use-workflow-approvals'
import {
  getApp,
  getAppBranch,
  getAppConfigs,
  getInstall,
  getInstallStack,
} from '@/lib'
import { humanize } from '@/utils/string-utils'
import { getWorkflowStepTitle } from '@/utils/workflow-utils'
import { OrgStatusBar } from './OrgStatusBar'
import { useInstallLink } from '@/hooks/use-install-path'

export const OrgStatusBarContainer = () => {
  const installLink = useInstallLink()
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

  const workflowItems: TContextTooltipItem[] = activeWorkflows.map((workflow) => ({
    id: workflow.id ?? '',
    title: workflow.name || humanize(workflow.type),
    subtitle: workflow.metadata?.owner_name || workflow.status?.status || undefined,
    href: workflow.owner_id
      ? installLink({ orgId: org.id, installId: workflow.owner_id, suffix: `/workflows/${workflow.id}` })
      : undefined,
    leftContent: (
      <Status
        status={workflow.status?.status ?? ''}
        isWithoutText
        variant="timeline"
        iconSize={16}
      />
    ),
  }))

  const ownerNames = new Map(
    activeWorkflows
      .filter((w) => w.owner_id && w.metadata?.owner_name)
      .map((w) => [w.owner_id!, w.metadata!.owner_name!])
  )

  const approvalItems: TContextTooltipItem[] = approvals.map((approval) => {
    const step = approval.workflow_step
    const href =
      step?.owner_id && step?.install_workflow_id
        ? installLink({ orgId: org.id, installId: step.owner_id, suffix: `/workflows/${step.install_workflow_id}` })
        : undefined
    const installName = step?.owner_id ? ownerNames.get(step.owner_id) : undefined
    return {
      id: approval.id ?? '',
      title: step?.name ? getWorkflowStepTitle(step) : 'Approval required',
      subtitle: installName || approval.type || undefined,
      href,
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
