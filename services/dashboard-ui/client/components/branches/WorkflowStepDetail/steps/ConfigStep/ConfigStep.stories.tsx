export default {
  title: 'Branches/WorkflowStepDetail/ConfigStep',
}

import { ConfigStep } from './ConfigStep'
import type { DiffSectionData } from './lib'

const sections: DiffSectionData[] = [
  {
    name: 'Components',
    additions: 1,
    removals: 0,
    changed: 1,
    entries: [
      { op: 'add', name: 'worker', description: 'new docker_build component' },
      { op: 'change', name: 'api', description: 'image tag updated' },
    ],
  },
  {
    name: 'Actions',
    additions: 0,
    removals: 1,
    changed: 0,
    entries: [{ op: 'remove', name: 'seed-db', description: 'action removed' }],
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
