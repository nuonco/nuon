export default {
  title: 'Installs/InstallConfiguration',
}

import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import type { TAppBranchRun, TAppInput } from '@/types'
import {
  InstallConfigurationAppBranch,
  InstallConfigurationConfigFile,
  InstallConfigurationInputs,
  InstallConfigurationOverrides,
} from './InstallConfiguration'

const appliedRun = {
  id: 'run-1',
  status: 'success',
  created_at: '2026-09-20T12:00:00Z',
  head_sha: 'a1b2c3d4e5f6',
  vcs_connection_commit: {
    sha: 'a1b2c3d4e5f6',
    message: 'Configure the payments worker',
    author_name: 'Example Developer',
  },
} as TAppBranchRun

const latestRun = {
  ...appliedRun,
  id: 'run-2',
  status: 'success',
  created_at: '2026-09-22T12:00:00Z',
  head_sha: 'f6e5d4c3b2a1',
  vcs_connection_commit: {
    sha: 'f6e5d4c3b2a1',
    message: 'Increase the checkout replica count',
    author_name: 'Example Developer',
  },
} as TAppBranchRun

const inputs = [
  {
    id: 'input-1',
    name: 'cluster_size',
    display_name: 'Cluster size',
    default: '3',
  },
  {
    id: 'input-2',
    name: 'domain',
    display_name: 'Domain',
    default: 'example.com',
  },
] as TAppInput[]

const hex = (value: string) =>
  Array.from(new TextEncoder().encode(value))
    .map((byte) => byte.toString(16).padStart(2, '0'))
    .join('')

const overrideInput = {
  id: 'override-1',
  name: `nuon_component_override_v1_helm_values_${hex('checkout')}`,
  display_name: 'Checkout Helm values',
} as TAppInput

export const AppBranchCurrent = () => (
  <InstallConfigurationAppBranch
    appliedConfigId="appcfg-15"
    branchName="release"
    branchHref="#"
    branchConfig={{
      connected_github_vcs_config: {
        repo: 'acme/platform-configs',
        branch: 'main',
        directory: 'apps/payments',
      },
    }}
    latestRun={appliedRun}
    latestRunHref="#"
    appliedRun={appliedRun}
    appliedRunHref="#"
    history={<Text theme="neutral">Applied config timeline</Text>}
  />
)

export const AppBranchUpdateAvailable = () => (
  <InstallConfigurationAppBranch
    appliedConfigId="appcfg-14"
    branchName="release"
    branchHref="#"
    branchConfig={{
      public_git_vcs_config: {
        repo: 'acme/platform-configs',
        branch: 'main',
        directory: 'apps/payments',
      },
    }}
    latestRun={latestRun}
    latestRunHref="#"
    appliedRun={appliedRun}
    appliedRunHref="#"
    history={<Text theme="neutral">Applied config timeline</Text>}
  />
)

export const AppBranchEmpty = () => <InstallConfigurationAppBranch />

export const Inputs = () => (
  <InstallConfigurationInputs
    action={<Button variant="secondary">Edit inputs</Button>}
    groups={[
      {
        id: 'group-1',
        name: 'platform',
        display_name: 'Platform',
        description: 'Runtime configuration for the application.',
        app_inputs: inputs,
      },
    ]}
    values={{
      cluster_size: '5',
      domain: 'payments.example.com',
    }}
  />
)

export const InputsEmpty = () => <InstallConfigurationInputs groups={[]} />

export const Overrides = () => (
  <InstallConfigurationOverrides
    action={<Button variant="secondary">Edit inputs</Button>}
    inputs={[overrideInput]}
    values={{
      [overrideInput.name!]: 'replicaCount: 4',
    }}
  />
)

export const OverridesEmpty = () => (
  <InstallConfigurationOverrides inputs={[]} values={{}} />
)

export const ConfigFile = () => (
  <InstallConfigurationConfigFile
    action={<Button variant="secondary">Sync now</Button>}
    content={`version = "v1"

[install]
name = "acme-production"

[install.inputs]
cluster_size = "5"
domain = "payments.example.com"
`}
    filename="installs/acme-production.toml"
    history={<Text theme="neutral">Config sync timeline</Text>}
    isManagedByConfig
    latestVersionId="cfg-15"
    syncedAt="2026-09-22T12:00:00Z"
  />
)

export const ConfigFileLoading = () => (
  <InstallConfigurationConfigFile
    isManagedByConfig
    isLoading
    history={<Text theme="neutral">Config sync timeline</Text>}
  />
)

export const ConfigFileDashboardManaged = () => (
  <InstallConfigurationConfigFile isManagedByConfig={false} />
)
