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
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { OrgStatusBar } from './OrgStatusBar'

export const OrgStatusBarContainer = () => {
  const { org } = useOrg()
  const { byocName, byocColor, byocTextColor } = useConfig()
  const { approvals } = useWorkflowApprovals()
  const { activeWorkflows } = useActiveWorkflows()
  const { appId, branchId, installId } = useParams()

  const { data: app } = useQuery({
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
  const latestConfig = appConfigs?.[0]

  const { data: branch } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-branch', org.id, appId, branchId],
    queryFn: () => getAppBranch({ orgId: org.id, appId: appId!, branchId: branchId! }),
    enabled: !!appId && !!branchId,
  })

  const { data: install } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install', org.id, installId],
    queryFn: () => getInstall({ orgId: org.id, installId: installId! }),
    enabled: !!installId,
  })

  const { data: stack } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-stack', org.id, installId],
    queryFn: () => getInstallStack({ installId: installId!, orgId: org.id }),
    enabled: !!installId,
    refetchInterval: 30_000,
  })

  const workflowItems: TContextTooltipItem[] = activeWorkflows.map((workflow) => ({
    id: workflow.id ?? '',
    title: workflow.name || toSentenceCase(snakeToWords(workflow.type)),
    subtitle: workflow.metadata?.owner_name || workflow.status?.status || undefined,
    href: workflow.owner_id
      ? `/${org.id}/installs/${workflow.owner_id}/workflows/${workflow.id}`
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
        ? `/${org.id}/installs/${step.owner_id}/workflows/${step.install_workflow_id}`
        : undefined
    const installName = step?.owner_id ? ownerNames.get(step.owner_id) : undefined
    return {
      id: approval.id ?? '',
      title: step?.name ? toSentenceCase(step.name) : 'Approval required',
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
