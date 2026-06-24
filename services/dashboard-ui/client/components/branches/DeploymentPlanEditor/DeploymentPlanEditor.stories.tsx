export default {
  title: 'Branches/DeploymentPlanEditor',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeploymentPlanEditor } from './DeploymentPlanEditor'
import type { IInstallGroup } from './types'

const noop = () => {}

const installs = [
  { id: 'i1', name: 'acme-prod', labels: { tier: 'prod', region: 'us-east-1' } },
  { id: 'i2', name: 'globex-prod', labels: { tier: 'prod', region: 'us-west-2' } },
  { id: 'i3', name: 'initech-staging', labels: { tier: 'staging', region: 'eu-west-1' } },
  { id: 'i4', name: 'umbrella-dev', labels: { tier: 'dev' } },
] as any

const groups: IInstallGroup[] = [
  {
    id: 'group-1',
    name: 'Production',
    install_ids: ['i1', 'i2'],
    label_selector: null,
    selection_mode: 'manual',
    order: 0,
    max_parallel: 1,
    use_for_previews: false,
  },
]

export const WithGroups = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor
      initialGroups={groups}
      availableInstalls={installs}
      loadingInstalls={false}
      isSaving={false}
      onSave={noop}
      onCancel={noop}
    />
  </ModalStory>
)

export const WithUnassigned = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor
      initialGroups={groups}
      availableInstalls={installs}
      loadingInstalls={false}
      isSaving={false}
      onSave={noop}
      onCancel={noop}
    />
  </ModalStory>
)

export const NoGroups = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor
      initialGroups={[]}
      availableInstalls={installs}
      loadingInstalls={false}
      isSaving={false}
      onSave={noop}
      onCancel={noop}
    />
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
    />
  </ModalStory>
)

export const Saving = () => (
  <ModalStory label="Open deployment plan">
    <DeploymentPlanEditor
      initialGroups={groups}
      availableInstalls={installs}
      loadingInstalls={false}
      isSaving
      onSave={noop}
      onCancel={noop}
    />
  </ModalStory>
)
