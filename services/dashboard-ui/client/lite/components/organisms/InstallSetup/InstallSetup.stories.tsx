import { useState } from 'react'
import type { TAPIError } from '@/types'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import {
  InstallSetup,
  type IInstallSetup as IInstallSetupProps,
} from './InstallSetup'

export default {
  title: 'lite/organisms/InstallSetup',
}

const apps: IInstallSetupProps['apps'] = [
  {
    id: 'app_payments',
    name: 'Payments API',
    platform: 'aws',
    source: 'acme/payments',
    updatedLabel: 'synced 8 minutes ago',
  },
  {
    id: 'app_analytics',
    name: 'Analytics warehouse',
    platform: 'azure',
    source: 'acme/analytics',
    updatedLabel: 'synced yesterday',
  },
]

const branches: IInstallSetupProps['branches'] = [
  {
    id: 'branch_main',
    name: 'main',
    groups: [
      {
        id: 'group_production',
        name: 'Production',
        kind: 'labels',
        labels: { env: 'production', tier: 'critical' },
      },
      {
        id: 'group_preview',
        name: 'Preview',
        kind: 'wildcard',
        labels: { pull_request: '*' },
      },
    ],
  },
  {
    id: 'branch_release',
    name: 'release',
    groups: [],
  },
]

const inputs: IInstallSetupProps['inputs'] = [
  {
    name: 'hostname',
    label: 'Public hostname',
    description: 'Hostname used by the public ingress.',
    required: true,
    type: 'text',
    defaultValue: '',
  },
  {
    name: 'replicas',
    label: 'Replica count',
    type: 'number',
    defaultValue: '2',
  },
  {
    name: 'metrics_enabled',
    label: 'Enable metrics',
    type: 'boolean',
    defaultValue: true,
  },
]

export const Overview = () => (
  <ComponentDocs
    name="InstallSetup"
    tier="organism"
    summary="The form-driven install setup wizard from app selection through provisioning."
    use={[
      'Render through InstallSetupContainer on the install setup route.',
      'Keep API queries and the create mutation in the container.',
    ]}
    avoid={[
      'Do not submit story fixture values to the API.',
      'Do not make branch enrollment mandatory.',
    ]}
    rules={[
      'App selection reveals optional branch and install-group enrollment.',
      'Location choices and account fields follow the selected cloud platform.',
      'App inputs come from the selected active app config.',
      'Concrete install-group labels render as locked label rows.',
      'Persistence is opt in so fixture stories always start clean.',
      'Create install is the only primary action on the submission step.',
    ]}
    props={[
      {
        name: 'apps',
        type: 'IAppSelectItem[]',
        description: 'Apps available for install creation.',
      },
      {
        name: 'branches',
        type: 'IInstallSetupBranch[]',
        description: 'Selected app branches and install groups.',
      },
      {
        name: 'inputs',
        type: 'IInstallSetupInput[]',
        description: 'Vendor inputs from the selected app config.',
      },
      {
        name: 'platform',
        type: 'TCloudPlatform',
        description: 'Selected app cloud platform.',
      },
      {
        name: 'configurationKey',
        type: 'string',
        description: 'Resets input defaults when the config changes.',
      },
      {
        name: 'configurationLoading',
        type: 'boolean',
        default: 'false',
        description: 'Loads app inputs.',
      },
      {
        name: 'configurationReady',
        type: 'boolean',
        default: 'false',
        description: 'Allows install creation.',
      },
      {
        name: 'appsLoading',
        type: 'boolean',
        default: 'false',
        description: 'Loads the app selector.',
      },
      {
        name: 'requireTargetAccount',
        type: 'boolean',
        default: 'false',
        description: 'Requires the target cloud account identifier.',
      },
      {
        name: 'pending',
        type: 'boolean',
        default: 'false',
        description: 'Shows install creation in progress.',
      },
      {
        name: 'error',
        type: 'unknown',
        description: 'Create-install API error shown in the form.',
      },
      {
        name: 'install',
        type: 'TInstall',
        description: 'Created install watched by the provision step.',
      },
      {
        name: 'persistence',
        type: '{ orgId: string; wizard: string }',
        description: 'Opts the form and wizard progress into local drafts.',
      },
      {
        name: 'onAppChange',
        type: '(appId: string) => void',
        description: 'Loads data for the selected app.',
      },
      {
        name: 'onBranchChange',
        type: '(branchId?: string) => void',
        description: 'Loads the selected branch config.',
      },
      {
        name: 'onClearError',
        type: '() => void',
        description: 'Clears a create error before retrying.',
      },
      {
        name: 'onSubmit',
        type: '(values: ICreateInstallValues) => void',
        description: 'Creates the install.',
      },
    ]}
  />
)

const InteractiveStory = () => {
  const [appId, setAppId] = useState('')
  const [error, setError] = useState<TAPIError>()
  return (
    <InstallSetup
      apps={apps}
      appsLoading={false}
      branches={appId ? branches : []}
      inputs={inputs}
      platform={apps.find((app) => app.id === appId)?.platform ?? 'unknown'}
      configurationKey={appId || undefined}
      configurationReady={!!appId}
      error={error}
      onAppChange={setAppId}
      onBranchChange={() => {}}
      onClearError={() => setError(undefined)}
      onSubmit={() =>
        setError({
          error: 'unable to create install: duplicated key not allowed',
          description: 'duplicate key',
          user_error: true,
          status: 409,
        })
      }
    />
  )
}

export const Interactive = () => <InteractiveStory />

export const LoadingApps = () => (
  <InstallSetup
    apps={[]}
    branches={[]}
    inputs={[]}
    platform="unknown"
    appsLoading
    onAppChange={() => {}}
    onBranchChange={() => {}}
    onSubmit={() => {}}
  />
)
