export default {
  title: 'Stacks/InstallStack',
}

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import {
  InstallStack,
  type IInstallStack,
  type TInstallNestedStack,
} from './InstallStack'

const baseProps: IInstallStack = {
  configVersion: 12,
  stackName: 'aws-eks',
  stackType: 'aws-cloudformation',
  nestedStacks: [],
  outputs: `{
  "vpc_id": "vpc-000000000000",
  "cluster_name": "acme-production",
  "region": "us-east-1"
}`,
  outputsAction: (
    <Button variant="secondary" size="sm">
      Trigger phone home
    </Button>
  ),
  configAction: <Button variant="secondary">Edit overrides</Button>,
  versionsAction: (
    <Button variant="secondary">
      <Icon variant="ClockCounterClockwiseIcon" size={16} />
      Stack versions
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

export const NoOutputs = () => (
  <InstallStack
    {...baseProps}
    nestedStacks={multipleNestedStacks}
    outputs={undefined}
    outputsAction={undefined}
  />
)

export const Loading = () => (
  <InstallStack
    {...baseProps}
    configLoading
    nestedStacks={[]}
    outputs={undefined}
    outputsLoading
  />
)

export const ConfigError = () => (
  <InstallStack
    {...baseProps}
    configError="Unable to load stack config."
    nestedStacks={[]}
  />
)
