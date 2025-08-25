import type {
  TInstall,
  TInstallActionWorkflow,
  TInstallActionWorkflowRun,
  TInstallComponent,
  TInstallComponentOutputs,
  TInstallDeploy,
  TInstallEvent,
  TInstallInputs,
  TInstallWorkflow,
  TInstallWorkflowStep,
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
  componentId: string
}

export async function getInstallComponent({
  componentId,
  installId,
  orgId,
}: IGetInstallComponent) {
  return queryData<TInstallComponent>({
    errorMessage: 'Unable to retrieve the components for this install.',
    orgId,
    path: `installs/${installId}/components/${componentId}`,
  })
}

export interface IGetInstallComponentDeploys extends IGetInstallComponent {}

export async function getInstallComponentDeploys({
  componentId,
  installId,
  orgId,
}: IGetInstallComponentDeploys) {
  return queryData<Array<TInstallDeploy>>({
    errorMessage: 'Unable to retrieve deployments for this install component.',
    orgId,
    path: `installs/${installId}/components/${componentId}/deploys`,
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
    abortTimeout: 100000,
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
    path: `installs/${installId}/action-workflows/runs`,
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
  return queryData<TSandboxRun>({
    errorMessage: 'Unable to retrieve install sandbox run.',
    orgId,
    path: `installs/sandbox-runs/${installSandboxRunId}`,
    abortTimeout: 10000,
  })
}

export interface IReprovisionInstall extends IGetInstall {}

export async function reprovisionInstall({
  installId,
  orgId,
}: IReprovisionInstall) {
  return mutateData({
    data: { error_behavior: 'string' },
    errorMessage: 'Unable to reprovision install.',
    orgId,
    path: `installs/${installId}/reprovision`,
  })
}

export async function reprovisionSandbox({
  installId,
  orgId,
}: IReprovisionInstall) {
  return mutateData({
    data: { error_behavior: 'string' },
    errorMessage: 'Unable to reprovision sandbox.',
    orgId,
    path: `installs/${installId}/reprovision-sandbox`,
  })
}

export interface IRunInstallActionWorkflow extends IGetInstall {
  actionWorkflowConfigId: string
  options?: {
    run_env_vars: Record<string, string>
  }
}

export async function runInstallActionWorkflow({
  actionWorkflowConfigId,
  installId,
  orgId,
  options,
}: IRunInstallActionWorkflow) {
  return mutateData({
    data: { action_workflow_config_id: actionWorkflowConfigId, ...options },
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
  return queryData<Array<TInstallActionWorkflow>>({
    errorMessage: 'Unable to retrieve latest install action workflow runs',
    orgId,
    path: `installs/${installId}/action-workflows/latest-runs`,
  })
}

export interface IGetInstallActionWorkflowRecentRuns extends IGetInstall {
  actionWorkflowId: string
  offset?: string;
  limit?: string
}

export async function getInstallActionWorkflowRecentRun({
  actionWorkflowId,
  installId,
  orgId,
}: IGetInstallActionWorkflowRecentRuns) {
  return queryData<TInstallActionWorkflow>({
    errorMessage: 'Unable to retrieve install action workflow runs',
    orgId,
    path: `installs/${installId}/action-workflows/${actionWorkflowId}/recent-runs`,
  })
}

export interface IDeployComponents extends IGetInstall {}

export async function deployComponents({
  installId,
  orgId,
}: IDeployComponents) {
  return mutateData({
    data: { error_behavior: 'string' },
    errorMessage: 'Unable to deploy components to install.',
    orgId,
    path: `installs/${installId}/components/deploy-all`,
  })
}

export interface IDeployComponentBuild extends IGetInstall {
  buildId: string
}

export async function deployComponentBuild({
  buildId,
  installId,
  orgId,
}: IDeployComponentBuild) {
  return mutateData({
    errorMessage: 'Unable to deploy component to install.',
    data: { build_id: buildId },
    orgId,
    path: `installs/${installId}/deploys`,
  })
}

export interface IGetInstallCurrentInputs extends IGetInstall {}

export async function getInstallCurrentInputs({
  installId,
  orgId,
}: IGetInstallCurrentInputs) {
  return queryData<TInstallInputs>({
    errorMessage: 'Unable to retrieve current install inputs.',
    orgId,
    path: `installs/${installId}/inputs/current`,
  })
}

export interface IGetInstallComponentOutputs extends IGetInstall {
  componentId: string
}

export async function getInstallComponentOutputs({
  componentId,
  installId,
  orgId,
}: IGetInstallComponentOutputs) {
  return queryData<TInstallComponentOutputs>({
    errorMessage: 'Unable to retrieve install component outputs.',
    orgId,
    path: `installs/${installId}/components/${componentId}/outputs`,
  })
}

export interface ICreateInstallData {
  name: string
  inputs?: Record<string, string>
  aws_account?: {
    iam_role_arn: string
    region: string
  }
  azure_account?: {
    location: string
    service_principal_app_id?: string
    service_principal_password?: string
    subscription_id?: string
    subscription_tenant_id?: string
  }
  metadata?: {
    managed_by?: string
  }
}

export const installManagedByUI = 'nuon/dashboard'

export interface ICreateInstall {
  appId: string
  orgId: string
  data: ICreateInstallData
}

export async function createInstall({ appId, orgId, data }: ICreateInstall) {
  return mutateData<TInstall>({
    errorMessage: 'Unable to create install.',
    data: data as unknown as Record<string, unknown>,
    orgId,
    path: `apps/${appId}/installs`,
  })
}

export interface ITeardownInstallComponents extends IGetInstall {}

export async function teardownInstallComponents({
  installId,
  orgId,
}: ITeardownInstallComponents) {
  return mutateData<string>({
    errorMessage: 'Unable to teardown install components.',
    orgId,
    path: `installs/${installId}/components/teardown-all`,
  })
}

export interface IUpdateInstall extends IGetInstall {
  data: {
    name: string
    inputs?: Record<string, string>
  }
}

export async function updateInstall({
  data,
  installId,
  orgId,
}: IUpdateInstall) {
  return mutateData<TInstall>({
    errorMessage: 'Unable to update install.',
    data,
    orgId,
    method: 'PATCH',
    path: `installs/${installId}`,
  })
}

export interface IForgetInstall extends IGetInstall {}

export async function forgetInstall({ installId, orgId }: IForgetInstall) {
  return mutateData<boolean>({
    errorMessage: 'Unable to forget install.',
    orgId,
    path: `installs/${installId}/forget`,
  })
}

export interface IGetInstallSandboxRuns extends IGetInstall {}

export async function getInstallSandboxRuns({
  installId,
  orgId,
}: IGetInstallSandboxRuns) {
  return queryData<Array<TSandboxRun>>({
    errorMessage: 'Unable to get install sandbox runs',
    orgId,
    path: `installs/${installId}/sandbox-runs`,
  })
}

export interface IGetInstallWorkflows extends IGetInstall {}

export async function getInstallWorkflows({
  installId,
  orgId,
}: IGetInstallWorkflows) {
  return queryData<Array<TInstallWorkflow>>({
    errorMessage: 'Unable to get install workflows.',
    path: `installs/${installId}/workflows`,
    orgId,
  })
}

export interface IGetInstallWorkflow {
  installWorkflowId: string
  orgId: string
}

export async function getInstallWorkflow({
  installWorkflowId,
  orgId,
}: IGetInstallWorkflow) {
  return queryData<TInstallWorkflow>({
    errorMessage: 'Unable to get install workflow.',
    path: `install-workflows/${installWorkflowId}`,
    orgId,
  })
}
