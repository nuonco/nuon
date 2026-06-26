export default {
  title: 'Branches/WorkflowStepDetail/ConfigStep',
}

import { ConfigStep } from './ConfigStep'
import type { DiffSectionData } from './lib'

const sections: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    grouped: true,
    additions: 1,
    removals: 0,
    changed: 1,
    entities: [
      {
        name: 'worker',
        op: 'add',
        componentType: 'docker_build',
        fields: [
          { key: 'type', op: 'add', diff: "'' -> 'docker_build'" },
          { key: 'dockerfile', op: 'add', diff: "'' -> 'Dockerfile'" },
        ],
      },
      {
        name: 'api',
        op: 'change',
        componentType: 'helm_chart',
        fields: [
          { key: 'image_tag', op: 'change', diff: "'v1.0' -> 'v1.1'" },
        ],
      },
    ],
    fields: [],
  },
  {
    name: 'Actions',
    sectionKey: 'actions',
    grouped: true,
    additions: 0,
    removals: 1,
    changed: 0,
    entities: [
      {
        name: 'seed-db',
        op: 'remove',
        fields: [
          { key: 'timeout', op: 'remove', diff: "'60s' -> ''" },
        ],
      },
    ],
    fields: [],
  },
]

export const WithChanges = () => (
  <ConfigStep
    appConfigId="cfg-123"
    status="success"
    sections={sections}
    summary={{ added: 1, removed: 1, changed: 1 }}
    diffResolved
    metadata={{}}
  />
)

export const NoChanges = () => (
  <ConfigStep
    appConfigId="cfg-123"
    status="success"
    sections={[]}
    summary={null}
    diffResolved
    metadata={{ component_count: 4, action_count: 2 }}
  />
)

export const LoadingDiff = () => (
  <ConfigStep
    appConfigId="cfg-123"
    status="success"
    sections={[]}
    summary={null}
    diffResolved={false}
    metadata={{}}
  />
)

export const WaitingForConfig = () => (
  <ConfigStep
    appConfigId={undefined}
    status="in-progress"
    sections={[]}
    summary={null}
    diffResolved={false}
    metadata={{}}
  />
)
