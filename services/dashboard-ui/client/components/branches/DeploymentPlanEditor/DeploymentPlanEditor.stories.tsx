export default {
  title: 'Branches/DeploymentPlanEditor',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeploymentPlanEditor } from './DeploymentPlanEditor'
import type { IInstallGroup } from './types'

const noop = () => {}

const installs = [
  {
    id: 'i1',
    name: 'acme-prod',
    labels: { tier: 'prod', region: 'us-east-1' },
  },
  {
    id: 'i2',
    name: 'globex-prod',
    labels: { tier: 'prod', region: 'us-west-2' },
  },
  {
    id: 'i3',
    name: 'initech-staging',
    labels: { tier: 'staging', region: 'eu-west-1' },
  },
  { id: 'i4', name: 'umbrella-dev', labels: { tier: 'dev' } },
] as any

const runbooks = [
  { id: 'rb1', name: 'db-migrate' },
  { id: 'rb2', name: 'smoke-test' },
  { id: 'rb3', name: 'warm-cache' },
] as any

const groups: IInstallGroup[] = [
  {
    id: 'group-1',
    name: 'Production',
    label_selector: { match_labels: { tier: 'prod' } },
    selection_mode: 'labels',
    is_default: false,
    order: 0,
    max_parallel: 1,
    auto_approve_on_policies_passing: false,
  },
  {
    id: 'group-default',
    name: 'Other installs',
    label_selector: null,
    selection_mode: 'pinned',
    is_default: true,
    order: 1,
    max_parallel: 1,
    auto_approve_on_policies_passing: false,
  },
]

const baseProps = {
  availableInstalls: installs,
  loadingInstalls: false,
  isSaving: false,
  onSave: noop,
  onCancel: noop,
  orgId: 'org123',
  runbooks,
  loadingRunbooks: false,
  initialPostDeployRunbookIds: [] as string[],
}

export const WithGroups = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor initialGroups={groups} {...baseProps} />
  </ModalStory>
)

export const MultipleGroups = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor
      initialGroups={[
        ...groups,
        {
          id: 'group-2',
          name: 'Manually assigned',
          label_selector: null,
          selection_mode: 'pinned',
          is_default: false,
          order: 2,
          max_parallel: 2,
          auto_approve_on_policies_passing: false,
        },
      ]}
      {...baseProps}
    />
  </ModalStory>
)

export const NoGroups = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor initialGroups={[]} {...baseProps} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor
      initialGroups={[]}
      availableInstalls={[]}
      loadingInstalls
      isSaving={false}
      onSave={noop}
      onCancel={noop}
      orgId="org123"
      runbooks={runbooks}
      loadingRunbooks={false}
      initialPostDeployRunbookIds={[]}
    />
  </ModalStory>
)

export const NoInstalls = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor
      initialGroups={[]}
      availableInstalls={[]}
      loadingInstalls={false}
      isSaving={false}
      onSave={noop}
      onCancel={noop}
      orgId="org123"
      runbooks={runbooks}
      loadingRunbooks={false}
      initialPostDeployRunbookIds={[]}
    />
  </ModalStory>
)

export const NoInstallsWithDefaultGroup = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor
      initialGroups={[
        {
          id: 'group-default',
          name: 'Everything',
          label_selector: null,
          selection_mode: 'pinned',
          is_default: true,
          order: 0,
          max_parallel: 1,
          auto_approve_on_policies_passing: false,
        },
      ]}
      availableInstalls={[]}
      loadingInstalls={false}
      isSaving={false}
      onSave={noop}
      onCancel={noop}
      orgId="org123"
      runbooks={runbooks}
      loadingRunbooks={false}
      initialPostDeployRunbookIds={[]}
    />
  </ModalStory>
)

export const DefaultGroup = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor
      initialGroups={[
        {
          id: 'group-default',
          name: 'Everything',
          label_selector: null,
          selection_mode: 'pinned',
          is_default: true,
          order: 0,
          max_parallel: 2,
          auto_approve_on_policies_passing: false,
        },
      ]}
      {...baseProps}
    />
  </ModalStory>
)

export const LabeledDefaultGroup = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor
      initialGroups={[
        {
          ...groups[0],
          id: 'group-labeled-default',
          name: 'Push installs',
          is_default: true,
        },
      ]}
      {...baseProps}
    />
  </ModalStory>
)

export const WithPostDeployRunbooks = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor
      initialGroups={groups}
      {...baseProps}
      initialPostDeployRunbookIds={['rb1', 'rb2']}
    />
  </ModalStory>
)

export const Saving = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor initialGroups={groups} {...baseProps} isSaving />
  </ModalStory>
)
