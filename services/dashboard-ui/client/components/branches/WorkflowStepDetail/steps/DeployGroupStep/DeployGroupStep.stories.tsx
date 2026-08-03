export default {
  title: 'Branches/WorkflowStepDetail/DeployGroupStep',
}

import { DeployGroupStep } from './DeployGroupStep'
import { InstallDeployRow, type IInstallDeployRow } from './InstallDeployRow'

const mkInstall = (over: Record<string, any> = {}): any => ({
  id: 'ins_acme',
  name: 'acme-prod',
  cloud_platform: 'aws',
  install_number: 12,
  aws_account: { region: 'us-east-1' },
  ...over,
})

const azureInstall = mkInstall({
  id: 'ins_globex',
  name: 'globex-staging',
  cloud_platform: 'azure',
  aws_account: undefined,
  azure_account: { location: 'eastus' },
})

const gcpInstall = mkInstall({
  id: 'ins_initech',
  name: 'initech-prod',
  cloud_platform: 'gcp',
  aws_account: undefined,
  gcp_account: { region: 'us-central1' },
})

const wfHref = '/org_1/installs/ins_acme/workflows/wf_1'
const installHref = '/org_1/installs/ins_acme'

export const RowDeployed = () => (
  <InstallDeployRow installId="ins_acme" install={mkInstall()} deployStatus="success" workflowHref={wfHref} installHref={installHref} />
)

export const RowInProgress = () => (
  <InstallDeployRow installId="ins_globex" install={azureInstall} deployStatus="in-progress" workflowHref={wfHref} installHref={installHref} />
)

export const RowError = () => (
  <InstallDeployRow installId="ins_initech" install={gcpInstall} deployStatus="error" workflowHref={wfHref} installHref={installHref} />
)

export const RowPending = () => (
  <InstallDeployRow installId="ins_acme" install={mkInstall({ cloud_platform: undefined, aws_account: undefined })} deployStatus="pending" installHref={installHref} />
)

export const RowUnresolved = () => (
  <InstallDeployRow installId="inlyompj5ren1oqpnvsc3xcksn" deployStatus="in-progress" workflowHref={wfHref} installHref={installHref} />
)

const rows: IInstallDeployRow[] = [
  { installId: 'ins_acme', install: mkInstall(), deployStatus: 'success', workflowHref: wfHref, installHref },
  { installId: 'ins_globex', install: azureInstall, deployStatus: 'in-progress', workflowHref: wfHref, installHref },
  { installId: 'ins_initech', install: gcpInstall, deployStatus: 'error', workflowHref: wfHref, installHref },
]

export const GroupMixed = () => (
  <DeployGroupStep groupName="UAT" totalInstalls={3} deployedCount={1} rows={rows} />
)

export const GroupAllDeployed = () => (
  <DeployGroupStep
    groupName="production"
    totalInstalls={2}
    deployedCount={2}
    rows={rows.slice(0, 2).map((r) => ({ ...r, deployStatus: 'success' }))}
  />
)

export const GroupSingleInstall = () => (
  <DeployGroupStep groupName="UAT" totalInstalls={1} deployedCount={1} rows={[rows[0]]} />
)

export const GroupDeploying = () => (
  <DeployGroupStep groupName="UAT" totalInstalls={0} deployedCount={0} rows={[]} emptyMessage="Deploying to install group" />
)
