export default {
  title: 'Stacks/InstallStack',
}

import { Button } from '@/components/common/Button'
import type { TInstallStackVersion } from '@/components/stacks/InstallStackVersions'
import {
  InstallStack,
  type IInstallStack,
  type TInstallNestedStack,
} from './InstallStack'

const versions = [
  {
    id: 'stkv-1',
    app_config_id: 'cfg-3',
    created_at: '2026-09-20T18:04:00Z',
    composite_status: { status: 'active' },
    runs: [{ id: 'run-1' }],
  },
  {
    id: 'stkv-2',
    app_config_id: 'cfg-2',
    created_at: '2026-09-18T12:00:00Z',
    composite_status: { status: 'active' },
    runs: [{ id: 'run-2' }],
  },
] as unknown as TInstallStackVersion[]

const baseProps: IInstallStack = {
  configVersion: 12,
  stackName: 'aws-eks',
  stackType: 'aws-cloudformation',
  nestedStacks: [],
  versions,
  configAction: <Button variant="secondary">Edit overrides</Button>,
  latestVersionAction: (
    <Button variant="secondary" size="sm">
      Reprovision stack
    </Button>
  ),
}

const multipleNestedStacks: TInstallNestedStack[] = [
  {
    id: 'runner',
    name: 'Runner',
    type: 'runner',
    source: 'app config',
    stack: {
      name: 'Runner',
      template_url: 'https://example.com/templates/runner.yaml',
    },
  },
  {
    id: 'vpc',
    name: 'VPC',
    type: 'vpc',
    source: 'install override',
    stack: {
      name: 'VPC',
      template_url: 'https://example.com/templates/custom-vpc.yaml',
    },
  },
  {
    id: 'custom-cluster',
    name: 'cluster',
    type: 'custom',
    source: 'app config',
    stack: {
      name: 'cluster',
      index: 0,
      template_url: './stacks/cluster.yaml',
      template_source_url: 'https://example.com/templates/cluster.yaml',
    },
  },
  {
    id: 'custom-database',
    name: 'database',
    type: 'custom',
    source: 'install only',
    stack: {
      name: 'database',
      index: 1,
      template_url: 'https://example.com/templates/database.yaml',
    },
  },
]

export const Default = () => <InstallStack {...baseProps} />

export const MultipleNestedStacks = () => (
  <InstallStack {...baseProps} nestedStacks={multipleNestedStacks} />
)

export const InstallOverrides = () => (
  <InstallStack
    {...baseProps}
    nestedStacks={multipleNestedStacks.map((stack) => ({
      ...stack,
      source: 'install override',
    }))}
  />
)

export const NoVersions = () => (
  <InstallStack
    {...baseProps}
    nestedStacks={multipleNestedStacks}
    versions={[]}
    latestVersionAction={undefined}
  />
)

export const Loading = () => (
  <InstallStack
    {...baseProps}
    configLoading
    nestedStacks={[]}
    versions={[]}
    versionsLoading
  />
)

export const ConfigError = () => (
  <InstallStack
    {...baseProps}
    configError="Unable to load stack config."
    nestedStacks={[]}
  />
)
