export default {
  title: 'Branches/DeploymentPlanEditor/GroupEditor',
}

import { GroupEditor } from './GroupEditor'
import type { IInstallGroup } from './types'

const noop = () => {}

const installs = [
  { id: 'i1', name: 'acme-prod', labels: { tier: 'prod', region: 'us-east-1' } },
  { id: 'i2', name: 'globex-prod', labels: { tier: 'prod', region: 'us-west-2' } },
  { id: 'i3', name: 'initech-staging', labels: { tier: 'staging', region: 'eu-west-1' } },
] as any

const manualGroup: IInstallGroup = {
  id: 'group-1',
  name: 'Production',
  install_ids: ['i1', 'i2'],
  label_selector: null,
  selection_mode: 'manual',
  order: 0,
  max_parallel: 1,
  use_for_previews: false,
}

const labelGroup: IInstallGroup = {
  id: 'group-2',
  name: 'Canary',
  install_ids: [],
  label_selector: { match_labels: { tier: 'staging' } },
  selection_mode: 'labels',
  order: 1,
  max_parallel: 2,
  use_for_previews: false,
}

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div className="max-w-2xl">{children}</div>
)

export const ManualSelection = () => (
  <Wrap>
    <GroupEditor
      group={manualGroup}
      index={0}
      totalGroups={2}
      availableInstalls={installs}
      unassignedInstalls={[installs[2]]}
      onUpdate={noop}
      onAddInstalls={noop}
      onRemoveInstall={noop}
      onMoveUp={noop}
      onMoveDown={noop}
      onDelete={noop}
    />
  </Wrap>
)

export const LabelSelector = () => (
  <Wrap>
    <GroupEditor
      group={labelGroup}
      index={1}
      totalGroups={2}
      availableInstalls={installs}
      unassignedInstalls={installs}
      onUpdate={noop}
      onAddInstalls={noop}
      onRemoveInstall={noop}
      onMoveUp={noop}
      onMoveDown={noop}
      onDelete={noop}
    />
  </Wrap>
)

export const EmptyManual = () => (
  <Wrap>
    <GroupEditor
      group={{ ...manualGroup, install_ids: [] }}
      index={0}
      totalGroups={1}
      availableInstalls={installs}
      unassignedInstalls={installs}
      onUpdate={noop}
      onAddInstalls={noop}
      onRemoveInstall={noop}
      onMoveUp={noop}
      onMoveDown={noop}
      onDelete={noop}
    />
  </Wrap>
)

export const WithNameError = () => (
  <Wrap>
    <GroupEditor
      group={{ ...manualGroup, name: '' }}
      index={0}
      totalGroups={1}
      availableInstalls={installs}
      unassignedInstalls={[installs[2]]}
      nameError="Group name is required"
      onUpdate={noop}
      onAddInstalls={noop}
      onRemoveInstall={noop}
      onMoveUp={noop}
      onMoveDown={noop}
      onDelete={noop}
    />
  </Wrap>
)
