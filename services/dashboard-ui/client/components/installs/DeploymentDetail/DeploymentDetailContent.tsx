import { useMemo, useState, type ReactNode } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import {
  changedBuildRows,
  type TBuildMeta,
} from '@/components/branches/BranchOverview/changed-builds'
import { BuildChangeCards } from '@/components/branches/BranchOverview/RunBuildsPanel'
import { BranchRunChanges } from '@/components/branches/BranchRunChanges/BranchRunChanges'
import { BranchRunChangesSummary } from '@/components/branches/BranchRunChangesSummary'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { ToggleButton } from '@/components/common/ToggleButton'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Text } from '@/components/common/Text'
import { WorkflowChangesSummaryContainer } from '@/components/workflows/WorkflowChangesSummary'
import { WorkflowAlertBanners } from '@/components/workflows/WorkflowDetails'
import { WorkflowSteps } from '@/components/workflows/WorkflowSteps'
import { PolicyViolations } from '@/components/workflows/step-details/PolicyViolations'
import { WorkflowActionButtons } from '@/components/workflows/workflow-details/WorkflowActionButtons'
import { useOrg } from '@/hooks/use-org'
import { useRespondedApprovals } from '@/hooks/use-responded-approvals'
import { useWorkflow } from '@/hooks/use-workflow'
import {
  getAppConfigDiff,
  getBranchRunBuilds,
  getBranchRunComparison,
  getBranchWorkflowRun,
} from '@/lib'
import { AppProvider } from '@/providers/app-provider'
import type { TInstallDeploymentRecord, TWorkflow } from '@/types'
import { getPolicyViolationCounts } from '@/utils/workflow-utils'
import { DeploymentConfigChanges } from './DeploymentConfigChanges'
import {
  deploymentChangeDescription,
  deploymentOutcomes,
  deploymentSteps,
  isAwaitingDeploymentApproval,
} from './deployment-progress'

const isBuildStep = (name?: string) =>
  !!name && /build/i.test(name) && !/(?:fetch|sync) app config/i.test(name)

const NO_BUILDS: TBuildMeta[] = []

const TemplateUpdateBlock = ({
  id,
  title,
  children,
  flush = false,
  actions,
  isOpen = false,
}: {
  id: string
  title: string
  children: ReactNode
  flush?: boolean
  actions?: ReactNode
  isOpen?: boolean
}) => (
  <Expand
    id={id}
    heading={
      <div className="flex flex-1 items-center justify-between gap-3">
        <Text variant="h3" weight="strong">
          {title}
        </Text>
        {actions ? (
          <span onClick={(event) => event.stopPropagation()}>{actions}</span>
        ) : null}
      </div>
    }
    isOpen={isOpen}
    interactiveHeading={!!actions}
    headerToggle={!!actions}
    className="rounded-xl border"
    headerClassName="items-center px-5 py-4"
  >
    <div className={flush ? 'border-t' : 'border-t p-4'}>{children}</div>
  </Expand>
)

type TTemplateView = 'template' | 'source'

const TemplateViewToggle = ({
  view,
  onChange,
}: {
  view: TTemplateView
  onChange: (view: TTemplateView) => void
}) => (
  <ToggleButton
    value={view}
    onChange={onChange}
    label="Template updates view"
    options={[
      {
        value: 'template',
        label: (
          <>
            <Icon variant="ListBulletsIcon" size={14} />
            List
          </>
        ),
      },
      {
        value: 'source',
        label: (
          <>
            <Icon variant="TreeStructureIcon" size={14} />
            Tree
          </>
        ),
      },
    ]}
  />
)

export const DeploymentTemplateContent = ({
  deployment,
  appId,
  workflow,
  isLoading = false,
  isError = false,
  onRetry,
}: {
  deployment?: TInstallDeploymentRecord
  appId?: string
  workflow?: TWorkflow
  isLoading?: boolean
  isError?: boolean
  onRetry?: () => void
}) => {
  const { org } = useOrg()
  const [view, setView] = useState<TTemplateView>('template')
  const branch = deployment?.app_branch
  const ready = !!branch?.id && !!branch.run_id && !!appId && !!org?.id
  const { data: branchRun } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'branch-workflow-run',
      org?.id,
      appId,
      branch?.id,
      branch?.run_id,
    ],
    queryFn: () =>
      getBranchWorkflowRun({
        orgId: org!.id,
        appId: appId!,
        branchId: branch!.id,
        runId: branch!.run_id!,
      }),
    enabled: ready && !isLoading && !isError,
  })
  const { data: runBuilds } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-run-builds', org?.id, appId, branch?.id, branch?.run_id],
    queryFn: () =>
      getBranchRunBuilds({
        orgId: org!.id,
        appId: appId!,
        branchId: branch!.id,
        runId: branch!.run_id!,
      }),
    enabled: ready && !isLoading && !isError,
  })
  // The template blocks start collapsed and unmount their content, so warm
  // the caches their queries read before the user expands them.
  const comparisonKey = [
    'branch-run-comparison',
    org?.id,
    appId,
    branch?.id,
    branch?.run_id,
  ]
  const comparisonArgs = {
    orgId: org?.id ?? '',
    appId: appId ?? '',
    branchId: branch?.id ?? '',
    runId: branch?.run_id ?? '',
  }
  useQuery({
    queryKey: comparisonKey,
    queryFn: () =>
      getBranchRunComparison({
        ...comparisonArgs,
        includeDiff: ['config', 'source'],
      }),
    enabled: ready && !isLoading && !isError,
    retry: 1,
  })
  const { data: configComparison } = useQuery({
    queryKey: [...comparisonKey, 'config'],
    queryFn: () =>
      getBranchRunComparison({ ...comparisonArgs, includeDiff: ['config'] }),
    enabled: ready && !isLoading && !isError,
    retry: 1,
  })
  const headConfigId = configComparison?.head_run?.app_config_id
  const baseConfigId = configComparison?.base_run?.app_config_id
  useQuery({
    queryKey: ['app-config-diff', org?.id, appId, headConfigId, baseConfigId],
    queryFn: () =>
      getAppConfigDiff({
        orgId: org!.id,
        appId: appId!,
        configId: headConfigId!,
        oldConfigId: baseConfigId,
      }),
    enabled: !!org?.id && !!appId && !!headConfigId,
    retry: 1,
  })
  const scope = useMemo(() => {
    const resources = deployment?.affected_resources
    const components = [
      ...new Set([
        ...(resources?.components ?? []),
        ...deploymentOutcomes(deployment, workflow)
          .filter((outcome) => outcome.category === 'Components')
          .map((outcome) => outcome.name),
      ]),
    ]
    return {
      sections: [
        ...(resources?.stack ? ['Stack', 'Permissions', 'Policies'] : []),
        ...(resources?.sandbox ? ['Sandbox'] : []),
        'Actions',
        ...(components.length ? ['Components'] : []),
        ...(deployment?.type === 'install_config_update'
          ? ['Install inputs', 'Secrets']
          : []),
      ],
      components,
    }
  }, [deployment, workflow])
  const buildStep = branchRun?.steps?.find((step) => isBuildStep(step.name))
  const buildMetadata = (buildStep?.status?.metadata ?? {}) as {
    builds?: TBuildMeta[]
    sandbox_build_id?: string
  }
  const metaBuilds = useMemo(
    () =>
      (buildMetadata.builds ?? NO_BUILDS).filter((build) =>
        build.component_type === 'sandbox' || build.component_id === 'sandbox'
          ? scope.sections.includes('Sandbox')
          : scope.components.includes(
              build.component_name || build.component_id || ''
            )
      ),
    [buildMetadata.builds, scope]
  )
  const changedBuilds = useMemo(
    () =>
      changedBuildRows({
        metaBuilds,
        runBuilds: (runBuilds ?? [])
          .filter((build) =>
            scope.components.includes(
              build.component_name || build.component_id || ''
            )
          )
          .map((build) => ({
            id: build.id,
            component_id: build.component_id,
            component_name: build.component_name,
            status: build.status_v2?.status || build.status,
          })),
        orgId: org?.id,
        appId,
        sandboxBuildId: buildMetadata.sandbox_build_id,
      }),
    [
      metaBuilds,
      runBuilds,
      scope,
      org?.id,
      appId,
      buildMetadata.sandbox_build_id,
    ]
  )
  if (isLoading) return <Text loading loadingWidth={32} />
  if (isError)
    return (
      <EmptyState
        emptyTitle="Template updates failed to load"
        emptyMessage="Unable to load this deployment's changes."
        action={
          onRetry ? (
            <Button variant="secondary" onClick={onRetry}>
              Try again
            </Button>
          ) : undefined
        }
      />
    )
  return branch?.id && branch.run_id && appId ? (
    <AppProvider appId={appId}>
      <div className="flex flex-col gap-4">
        <TemplateUpdateBlock
          id="deployment-template-updates"
          title="Template updates"
          isOpen
          flush
          actions={<TemplateViewToggle view={view} onChange={setView} />}
        >
          {view === 'template' ? (
            <BranchRunChangesSummary
              branchId={branch.id}
              appBranchRunId={branch.run_id}
              builds={metaBuilds}
              scope={scope}
              bare
            />
          ) : (
            <BranchRunChanges
              branchId={branch.id}
              appBranchRunId={branch.run_id}
              showRunComparison={false}
              showConfigChanges={false}
              bare
              scope={scope}
            />
          )}
        </TemplateUpdateBlock>
        <TemplateUpdateBlock id="deployment-builds" title="Builds">
          <BuildChangeCards rows={changedBuilds} bare />
        </TemplateUpdateBlock>
      </div>
    </AppProvider>
  ) : (
    <DeploymentConfigChanges deployment={deployment} />
  )
}

export const DeploymentAlerts = () => {
  const { workflow, failedSteps } = useWorkflow()
  const { hasResponded } = useRespondedApprovals()
  const steps = deploymentSteps(workflow)
  const approvals = steps.filter(
    (step) =>
      isAwaitingDeploymentApproval(step) &&
      step.approval?.type &&
      !['approve-all', 'noop'].includes(step.approval.type) &&
      !hasResponded(step.id)
  )
  return (
    <>
      <WorkflowAlertBanners workflow={workflow} failedSteps={failedSteps} />
      {approvals.map((step) => (
        <ApprovalBanner key={step.id} step={step} />
      ))}
      {steps
        .filter(
          (step) =>
            getPolicyViolationCounts(step).hasViolations &&
            !failedSteps.some((failed) => failed.id === step.id)
        )
        .map((step) => (
          <div key={step.id} className="flex shrink-0 flex-col gap-2">
            <Text variant="subtext" weight="strong">
              {step.name}
            </Text>
            <PolicyViolations step={step} />
          </div>
        ))}
    </>
  )
}

export const DeploymentWorkflowContent = ({
  showChangesTrigger = false,
}: {
  showChangesTrigger?: boolean
}) => {
  const { workflow } = useWorkflow()
  return (
    <div className="flex flex-col gap-4">
      <SectionHeader
        title="Workflow steps"
        description={
          workflow.status?.status === 'error'
            ? undefined
            : workflow.status?.status_human_description
        }
        actions={
          <WorkflowActionButtons showChangesTrigger={showChangesTrigger} />
        }
      />
      <WorkflowSteps
        approvalPrompt={workflow.approval_option === 'prompt'}
        planOnly={workflow.plan_only}
      />
    </div>
  )
}

export const DeploymentChangesContent = ({
  deployment,
}: {
  deployment?: TInstallDeploymentRecord
}) => {
  const { workflow } = useWorkflow()
  return (
    <div className="flex flex-col gap-4">
      <SectionHeader
        title="Change summary"
        description={deploymentChangeDescription(
          workflow,
          deploymentOutcomes(deployment, workflow)
        )}
      />
      <WorkflowChangesSummaryContainer />
    </div>
  )
}
