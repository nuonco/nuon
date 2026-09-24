export default {
  title: 'Branches/DeploymentPlanEditor/GroupEditor',
}

import { GroupEditor } from './GroupEditor'
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
] as any

const labelGroup: IInstallGroup = {
  id: 'group-1',
  name: 'Production',
  label_selector: { match_labels: { tier: 'prod' } },
  selection_mode: 'labels',
  is_default: false,
  order: 0,
  max_parallel: 1,
  auto_approve_on_policies_passing: false,
}

const defaultGroup: IInstallGroup = {
  id: 'group-2',
  name: 'Everything',
  label_selector: null,
  selection_mode: 'pinned',
  is_default: true,
  order: 1,
  max_parallel: 1,
  auto_approve_on_policies_passing: false,
}

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div className="max-w-3xl">{children}</div>
)

export const LabelSelector = () => (
  <Wrap>
    <GroupEditor
      group={labelGroup}
      index={0}
      totalGroups={2}
      availableInstalls={installs}
      onUpdate={noop}
      onMoveUp={noop}
      onMoveDown={noop}
      onDelete={noop}
    />
  </Wrap>
)

export const EmptyLabelSelector = () => (
  <Wrap>
    <GroupEditor
      group={{ ...labelGroup, label_selector: null }}
      index={0}
      totalGroups={1}
      availableInstalls={installs}
      onUpdate={noop}
      onMoveUp={noop}
      onMoveDown={noop}
      onDelete={noop}
    />
  </Wrap>
)

export const DefaultSelection = () => (
  <Wrap>
    <GroupEditor
      group={defaultGroup}
      index={1}
      totalGroups={2}
      availableInstalls={installs}
      onUpdate={noop}
      onMoveUp={noop}
      onMoveDown={noop}
      onDelete={noop}
    />
  </Wrap>
)

export const DefaultSelectionNoInstalls = () => (
  <Wrap>
    <GroupEditor
      group={defaultGroup}
      index={0}
      totalGroups={1}
      availableInstalls={[]}
      onUpdate={noop}
      onMoveUp={noop}
      onMoveDown={noop}
      onDelete={noop}
    />
  </Wrap>
)

export const LabeledDefaultSelection = () => (
  <Wrap>
    <GroupEditor
      group={{ ...labelGroup, is_default: true }}
      index={0}
      totalGroups={1}
      availableInstalls={installs}
      onUpdate={noop}
      onMoveUp={noop}
      onMoveDown={noop}
      onDelete={noop}
    />
  </Wrap>
)

export const WithNameError = () => (
  <Wrap>
    <GroupEditor
      group={{ ...labelGroup, name: '' }}
      index={0}
      totalGroups={1}
      availableInstalls={installs}
      nameError="Group name is required"
      onUpdate={noop}
      onMoveUp={noop}
      onMoveDown={noop}
      onDelete={noop}
    />
  </Wrap>
)
