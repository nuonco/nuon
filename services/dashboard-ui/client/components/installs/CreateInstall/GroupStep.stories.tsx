export default {
  title: 'Installs/GroupStep',
}

import { useState } from 'react'
import { GroupStep, type TGroupStepSelection } from './GroupStep'
import type { TAppBranchConfig } from '@/types'

const noop = () => {}

const mixedConfig: TAppBranchConfig = {
  id: 'cfg-1',
  install_groups: [
    {
      id: 'g-prod',
      name: 'Production',
      label_selector: {
        match_labels: { env: 'production', region: 'us-east-1' },
      },
    },
    {
      id: 'g-staging',
      name: 'Staging',
      label_selector: { match_labels: { env: 'staging' } },
    },
    {
      id: 'g-wildcard',
      name: 'All regions',
      label_selector: { match_labels: { region: '*' } },
    },
    {
      id: 'g-default',
      name: 'Remaining installs',
      default: true,
    },
  ],
} as unknown as TAppBranchConfig

const emptyConfig: TAppBranchConfig = {
  id: 'cfg-empty',
  install_groups: [],
} as unknown as TAppBranchConfig

const labelOnlyConfig: TAppBranchConfig = {
  id: 'cfg-labels',
  install_groups: [
    {
      id: 'g1',
      name: 'EMEA',
      label_selector: { match_labels: { region: 'eu-west-1' } },
    },
    {
      id: 'g2',
      name: 'US',
      label_selector: { match_labels: { region: 'us-east-1' } },
    },
    {
      id: 'g-default',
      name: 'Other installs',
      default: true,
    },
  ],
} as unknown as TAppBranchConfig

export const Default = () => {
  const [selected, setSelected] = useState<TGroupStepSelection>({
    mode: 'none',
    group: null,
  })
  return (
    <div className="p-6 max-w-lg">
      <GroupStep
        config={mixedConfig}
        selected={selected}
        onSelect={setSelected}
      />
    </div>
  )
}

export const LabelGroups = () => {
  const [selected, setSelected] = useState<TGroupStepSelection>({
    mode: 'none',
    group: null,
  })
  return (
    <div className="p-6 max-w-lg">
      <GroupStep
        config={labelOnlyConfig}
        selected={selected}
        onSelect={setSelected}
      />
    </div>
  )
}

export const LabelGroupSelected = () => {
  const [selected, setSelected] = useState<TGroupStepSelection>({
    mode: 'labels',
    group: labelOnlyConfig.install_groups![0],
  })
  return (
    <div className="p-6 max-w-lg">
      <GroupStep
        config={labelOnlyConfig}
        selected={selected}
        onSelect={setSelected}
      />
    </div>
  )
}

export const EmptyConfig = () => (
  <div className="p-6 max-w-lg">
    <GroupStep
      config={emptyConfig}
      selected={{ mode: 'none', group: null }}
      onSelect={noop}
    />
  </div>
)

export const DefaultGroup = () => {
  const defaultOnlyConfig: TAppBranchConfig = {
    id: 'cfg-no-labels',
    install_groups: [{ id: 'g-default', name: 'Everyone', default: true }],
  } as unknown as TAppBranchConfig

  return (
    <div className="p-6 max-w-lg">
      <GroupStep
        config={defaultOnlyConfig}
        selected={{
          mode: 'default',
          group: defaultOnlyConfig.install_groups![0],
        }}
        onSelect={noop}
      />
    </div>
  )
}

export const PinnedOnlyGroup = () => {
  const pinnedConfig: TAppBranchConfig = {
    id: 'cfg-pinned',
    install_groups: [
      { id: 'g-pinned', name: 'Canary' },
      { id: 'g-default', name: 'Other installs', default: true },
    ],
  } as unknown as TAppBranchConfig
  const [selected, setSelected] = useState<TGroupStepSelection>({
    mode: 'none',
    group: pinnedConfig.install_groups![0],
  })

  return (
    <div className="p-6 max-w-lg">
      <GroupStep
        config={pinnedConfig}
        selected={selected}
        onSelect={setSelected}
      />
    </div>
  )
}
