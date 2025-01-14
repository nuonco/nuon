import type {
  TActionWorkflow,
  TInstall,
  TInstallActionWorkflowRun,
  TInstallComponent,
  TInstallDeploy,
  TInstallDeployPlan,
  TInstallEvent,
  TReadme,
  TRunnerGroup,
  TSandboxRun,
} from '@/types'
import { mutateData, queryData } from '@/utils'

export interface IGetInstalls {
  orgId: string
}

export async function getInstalls({ orgId }: IGetInstalls) {
  return queryData<Array<TInstall>>({
    errorMessage: 'Unable to retrieve your installs.',
    orgId,
    path: 'installs',
  })
}

export interface IGetInstall extends IGetInstalls {
  installId: string
}

export async function getInstall({ installId, orgId }: IGetInstall) {
  return queryData<TInstall>({
    errorMessage: 'Unable to retrieve install.',
    orgId,
    path: `installs/${installId}`,
  })
}

export interface IGetInstallComponents extends IGetInstall {}

export async function getInstallComponents({
  installId,
  orgId,
}: IGetInstallComponents) {
  return queryData<Array<TInstallComponent>>({
    errorMessage: 'Unable to retrieve the components for this install.',
    orgId,
    path: `installs/${installId}/components`,
  })
}

export interface IGetInstallComponent extends IGetInstall {
  installComponentId: string
}

export async function getInstallComponent({
  installComponentId,
  installId,
  orgId,
}: IGetInstallComponent) {
  return (
    await queryData<Array<TInstallComponent>>({
      errorMessage: 'Unable to retrieve the components for this install.',
      orgId,
      path: `installs/${installId}/components`,
    })
  ).find((installComponent) => installComponent.id === installComponentId)
}

export interface IGetInstallComponentDeploys extends IGetInstallComponent {}

export async function getInstallComponentDeploys({
  installComponentId,
  installId,
  orgId,
}: IGetInstallComponentDeploys) {
  return queryData<Array<TInstallDeploy>>({
    errorMessage: 'Unable to retrieve deployments for this install component.',
    orgId,
    path: `installs/${installId}/components/${installComponentId}/deploys`,
  })
}

export interface IGetInstallDeploy extends IGetInstall {
  installDeployId: string
}

export async function getInstallDeploy({
  installDeployId,
  installId,
  orgId,
}: IGetInstallDeploy) {
  return queryData<TInstallDeploy>({
    errorMessage: 'Unable to retrieve install deployment.',
    orgId,
    path: `installs/${installId}/deploys/${installDeployId}`,
  })
}

export interface IGetInstallEvents extends IGetInstall {}

export async function getInstallEvents({
  installId,
  orgId,
}: IGetInstallEvents) {
  return queryData<Array<TInstallEvent>>({
    errorMessage: 'Unable to retrieve install events.',
    orgId,
    path: `installs/${installId}/events`,
  })
}

export interface IGetInstallReadme extends IGetInstall {}

export async function getInstallReadme({
  orgId,
  installId,
}: IGetInstallReadme) {
  return queryData<TReadme>({
    errorMessage: 'Unable to retrieve the install README.',
    orgId,
    path: `installs/${installId}/readme`,
  })
}

export interface IGetInstallRunnerGroup extends IGetInstall {}

export async function getInstallRunnerGroup({
  installId,
  orgId,
}: IGetInstallRunnerGroup) {
  return queryData<TRunnerGroup>({
    errorMessage: 'Unable to retrieve install runner group.',
    orgId,
    path: `installs/${installId}/runner-group`,
  })
}

export interface IGetInstallActionWorkflowRuns extends IGetInstall {}

export async function getInstallActionWorkflowRuns({
  installId,
  orgId,
}: IGetInstallActionWorkflowRuns) {
  return queryData<Array<TInstallActionWorkflowRun>>({
    errorMessage: 'Unable to retrieve install action workflow runs.',
    orgId,
    path: `installs/${installId}/action-workflows/runs/`,
  })
}

export interface IGetInstallActionWorkflowRun extends IGetInstall {
  actionWorkflowRunId: string
}

export async function getInstallActionWorkflowRun({
  actionWorkflowRunId,
  installId,
  orgId,
}: IGetInstallActionWorkflowRun) {
  return queryData<TInstallActionWorkflowRun>({
    errorMessage: 'Unable to retrieve install action workflow run.',
    orgId,
    path: `installs/${installId}/action-workflows/runs/${actionWorkflowRunId}`,
  })
}

export interface IGetInstallSandboxRun extends IGetInstall {
  installSandboxRunId: string
}

export async function getInstallSandboxRun({
  installSandboxRunId,
  installId,
  orgId,
}: IGetInstallSandboxRun) {
  return (
    await queryData<Array<TSandboxRun>>({
      errorMessage: 'Unable to retrieve install sandbox run.',
      orgId,
      path: `installs/${installId}/sandbox-runs`,
    })
  ).find((sandboxRun) => sandboxRun.id === installSandboxRunId)
}

export interface IReprovisionInstall extends IGetInstall {}

export async function reprovisionInstall({
  installId,
  orgId,
}: IReprovisionInstall) {
  return mutateData({
    errorMessage: 'Unable to reprovision install.',
    orgId,
    path: `installs/${installId}/reprovision`,
  })
}

export interface IRunInstallActionWorkflow extends IGetInstall {
  actionWorkflowConfigId: string
}

export async function runInstallActionWorkflow({
  actionWorkflowConfigId,
  installId,
  orgId,
}: IRunInstallActionWorkflow) {
  return mutateData({
    data: { action_workflow_config_id: actionWorkflowConfigId },
    errorMessage: 'Unable to run action workflow on this install.',
    orgId,
    path: `installs/${installId}/action-workflows/runs`,
  })
}

export interface IGetInstallActionWorkflowLatestRuns extends IGetInstall {}

export async function getInstallActionWorkflowLatestRun({
  installId,
  orgId,
}: IGetInstallActionWorkflowLatestRuns) {
  return queryData<
    Array<{
      action_workflow: TActionWorkflow
      install_action_workflow_run: TInstallActionWorkflowRun
    }>
  >({
    errorMessage: 'Unable to retrieve latest install action workflow runs',
    orgId,
    path: `installs/${installId}/action-workflows/latest-runs`,
  })
}

export interface IGetInstallActionWorkflowRecentRuns extends IGetInstall {
  actionWorkflowId: string
}

export async function getInstallActionWorkflowRecentRun({
  actionWorkflowId,
  installId,
  orgId,
}: IGetInstallActionWorkflowRecentRuns) {
  return queryData<{
    action_workflow: TActionWorkflow
    recent_runs: Array<TInstallActionWorkflowRun>
  }>({
    errorMessage: 'Unable to retrieve install action workflow runs',
    orgId,
    path: `installs/${installId}/action-workflows/${actionWorkflowId}/recent-runs`,
  })
}

export interface IDeployComponents extends IGetInstall {}

export async function deployComponents({
  installId,
  orgId,
}: IReprovisionInstall) {
  return mutateData({
    errorMessage: 'Unable to deploy components to install.',
    orgId,
    path: `installs/${installId}/components/deploy-all`,
  })
}

export interface IGetInstallDeployPlan extends IGetInstall {
  deployId: string
}

export async function getInstallDeployPlan({
  deployId,
  installId,
  orgId,
}: IGetInstallDeployPlan) {
  return queryData<TInstallDeployPlan>({
    errorMessage: 'Unable to retrieve install deplay plan.',
    orgId,
    path: `installs/${installId}/deploys/${deployId}/plan`,
  })
}
