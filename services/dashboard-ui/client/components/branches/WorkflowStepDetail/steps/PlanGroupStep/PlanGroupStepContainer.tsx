import { keepPreviousData, useQueries, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppConfigDiff } from '@/lib'
import {
  extractSections,
  computeSummary,
} from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import type { TInstallWorkflowStep } from '@/types'
import { PlanGroupStep, type PlanInstallDiff } from './PlanGroupStep'
import { GroupApprovalActions } from './GroupApprovalActions'

interface IPlanGroupStepContainer {
  step: TInstallWorkflowStep
  metadata: Record<string, any>
}

export const PlanGroupStepContainer = ({
  step,
  metadata,
}: IPlanGroupStepContainer) => {
  const { org } = useOrg()
  const { app, labelColors } = useApp()
  const orgId = org?.id ?? ''
  const appId = app?.id ?? ''

  const approvalId = step.approval?.id
  const hasApproval = step.execution_type === 'approval' && !!approvalId
  const hasResponse = !!step.approval?.response
  const isAwaiting = step.status?.status === 'approval-awaiting'

  const { data: plan } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['approval-plan', orgId, step.id, approvalId],
    queryFn: async () => {
      const res = await fetch(
        `/api/orgs/${orgId}/workflows/${step.install_workflow_id}/steps/${step.id}/approvals/${approvalId}/contents`
      )
      if (!res.ok)
        throw new Error(`Failed to fetch approval contents: ${res.status}`)
      return res.json()
    },
    enabled: !!orgId && !!step.id && !!step.install_workflow_id && !!approvalId,
  })

  const rawInstalls = (plan?.installs || metadata.installs || []) as any[]
  const groupName =
    plan?.install_group ||
    metadata.install_group_name ||
    step.name?.replace(/^plan install group:\s*/i, '')
  const showApproveBar = hasApproval && isAwaiting && !hasResponse

  const diffQueries = useQueries({
    queries: rawInstalls.map((inst) => ({
      queryKey: [
        'app-config-diff',
        orgId,
        appId,
        inst.new_app_config_id,
        inst.old_app_config_id,
      ],
      queryFn: () =>
        getAppConfigDiff({
          orgId,
          appId,
          configId: inst.new_app_config_id,
          oldConfigId: inst.old_app_config_id,
        }),
      enabled: !!orgId && !!appId && !!inst.new_app_config_id,
    })),
  })

  const installs: PlanInstallDiff[] = rawInstalls.map((inst, i) => {
    const query = diffQueries[i]
    const sections = query?.data?.diff ? extractSections(query.data.diff) : []
    const summary =
      sections.length > 0
        ? computeSummary(sections)
        : query?.data?.summary
          ? {
              added: query.data.summary.added,
              removed: query.data.summary.removed,
              changed: query.data.summary.changed,
            }
          : null

    return {
      installId: inst.install_id,
      installName: inst.install_name || inst.install_id,
      installLabels: inst.install_labels,
      sections,
      summary,
      isLoading: !!query?.isLoading,
    }
  })

  return (
    <PlanGroupStep
      installs={installs}
      groupName={groupName}
      labelColors={labelColors}
      hasResponse={hasResponse}
      responseType={step.approval?.response?.response_type}
      showApproveBar={showApproveBar}
      isInProgress={step.status?.status === 'in-progress'}
      actions={
        showApproveBar ? (
          <GroupApprovalActions
            target={{
              orgId,
              workflowId: step.install_workflow_id,
              stepId: step.id,
              approvalId: approvalId!,
              groupName: groupName || 'install group',
            }}
          />
        ) : undefined
      }
    />
  )
}
