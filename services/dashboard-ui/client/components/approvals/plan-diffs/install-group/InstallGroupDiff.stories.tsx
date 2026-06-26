export default {
  title: 'Approvals/PlanDiffs/InstallGroupDiff',
}

import { InstallGroupDiff } from './InstallGroupDiff'
import type { InstallDiffEntry } from './InstallGroupDiff'

const mockInstalls: InstallDiffEntry[] = [
  {
    installId: 'inst-abc123',
    installName: 'production-us-west-2',
    installLabels: { env: 'production', region: 'us-west-2' },
    status: 'pending',
    sandboxChanged: false,
    stackChanged: false,
    summary: { added: 1, removed: 0, changed: 2 },
    sections: [
      {
        name: 'Components',
        sectionKey: 'components',
        grouped: true,
        additions: 1,
        removals: 0,
        changed: 2,
        entities: [
          {
            name: 'redis',
            op: 'add',
            componentType: 'helm_chart',
            fields: [
              { key: 'type', op: 'add', diff: "'' -> 'helm_chart'" },
              { key: 'chart_name', op: 'add', diff: "'' -> 'redis'" },
            ],
          },
          {
            name: 'ctl-api',
            op: 'change',
            componentType: 'helm_chart',
            fields: [
              { key: 'image_tag', op: 'change', diff: "'v2.3' -> 'v2.4'" },
            ],
          },
          {
            name: 'dashboard',
            op: 'change',
            componentType: 'docker_build',
            fields: [
              { key: 'dockerfile', op: 'change', diff: "'Dockerfile' -> 'Dockerfile.prod'" },
            ],
          },
        ],
        fields: [],
      },
    ],
  },
  {
    installId: 'inst-def456',
    installName: 'staging-us-east-1',
    installLabels: { env: 'staging', region: 'us-east-1' },
    status: 'pending',
    sandboxChanged: true,
    stackChanged: false,
    summary: { added: 1, removed: 0, changed: 0 },
    sections: [
      {
        name: 'Components',
        sectionKey: 'components',
        grouped: true,
        additions: 1,
        removals: 0,
        changed: 0,
        entities: [
          {
            name: 'redis',
            op: 'add',
            componentType: 'helm_chart',
            fields: [
              { key: 'type', op: 'add', diff: "'' -> 'helm_chart'" },
            ],
          },
        ],
        fields: [],
      },
    ],
  },
  {
    installId: 'inst-ghi789',
    installName: 'dev-local',
    status: 'pending',
    summary: { added: 0, removed: 0, changed: 0 },
    sections: [],
  },
]

export const Default = () => (
  <InstallGroupDiff groupName="production" installs={mockInstalls} />
)

export const Empty = () => (
  <InstallGroupDiff groupName="staging" installs={[]} />
)

export const SingleInstall = () => (
  <InstallGroupDiff groupName="canary" installs={[mockInstalls[0]]} />
)

export const NoChangesInstall = () => (
  <InstallGroupDiff groupName="dev" installs={[mockInstalls[2]]} />
)
