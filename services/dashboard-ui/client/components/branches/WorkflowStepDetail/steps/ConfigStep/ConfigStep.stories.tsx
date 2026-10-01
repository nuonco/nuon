export default {
  title: 'Features / Branches / Workflow step detail / Config step',
}

import { ConfigStep } from './ConfigStep'
import type { DiffSectionData } from './lib'

const sections: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    grouped: true,
    additions: 2,
    removals: 0,
    changed: 0,
    entities: [
      {
        name: 'worker',
        op: 'add',
        componentType: 'docker_build',
        fields: [
          { key: 'type', op: 'add', diff: "'docker_build'" },
          { key: 'dockerfile', op: 'add', diff: "'Dockerfile'" },
        ],
      },
      {
        name: 'api',
        op: 'add',
        componentType: 'helm_chart',
        fields: [
          { key: 'chart_name', op: 'add', diff: "'api'" },
          { key: 'image_tag', op: 'add', diff: "'v1.1'" },
        ],
      },
    ],
    fields: [],
  },
  {
    name: 'Actions',
    sectionKey: 'actions',
    grouped: true,
    additions: 1,
    removals: 0,
    changed: 0,
    entities: [
      {
        name: 'seed-db',
        op: 'add',
        fields: [{ key: 'timeout', op: 'add', diff: "'60s'" }],
      },
    ],
    fields: [],
  },
]

export const FullConfig = () => (
  <ConfigStep appConfigId="cfg-123" status="success" sections={sections} />
)

export const Loading = () => (
  <ConfigStep appConfigId="cfg-123" status="success" sections={[]} isLoading />
)

export const Empty = () => (
  <ConfigStep appConfigId="cfg-123" status="success" sections={[]} />
)

export const ExcludedSectionsHidden = () => (
  <ConfigStep
    appConfigId="cfg-123"
    status="success"
    sections={[
      ...sections,
      {
        name: 'Stack',
        sectionKey: 'stack',
        grouped: false,
        additions: 1,
        removals: 0,
        changed: 0,
        entities: [],
        fields: [{ key: 'name', op: 'add', diff: "'acme-prod'" }],
        files: [{ name: 'stack.toml', op: 'add' }],
      },
      {
        name: 'Install inputs',
        sectionKey: 'inputs',
        grouped: true,
        additions: 1,
        removals: 0,
        changed: 0,
        entities: [
          {
            name: 'environment',
            op: 'add',
            fields: [{ key: 'default', op: 'add', diff: "'staging'" }],
          },
        ],
        fields: [],
      },
    ]}
  />
)

export const WaitingForConfig = () => (
  <ConfigStep appConfigId={undefined} status="in-progress" sections={[]} />
)

export const PendingConfig = () => (
  <ConfigStep appConfigId={undefined} status="pending" sections={[]} />
)
