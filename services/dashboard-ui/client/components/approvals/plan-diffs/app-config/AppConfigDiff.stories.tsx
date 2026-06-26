export default {
  title: 'Approvals/PlanDiffs/AppConfigDiff',
}

import { AppConfigDiff } from './AppConfigDiff'
import type { DiffSectionData } from './AppConfigDiff'

const mockSections: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    grouped: true,
    additions: 1,
    removals: 1,
    changed: 1,
    entities: [
      {
        name: 'redis',
        op: 'add',
        componentType: 'helm_chart',
        fields: [
          { key: 'type', op: 'add', diff: "'' -> 'helm_chart'" },
          { key: 'chart_name', op: 'add', diff: "'' -> 'redis'" },
          { key: 'namespace', op: 'add', diff: "'' -> 'cache'" },
        ],
        files: [],
      },
      {
        name: 'legacy-worker',
        op: 'remove',
        componentType: 'docker_build',
        fields: [
          { key: 'type', op: 'remove', diff: "'docker_build' -> ''" },
          { key: 'dockerfile', op: 'remove', diff: "'Dockerfile.worker' -> ''" },
        ],
        files: [],
      },
      {
        name: 'ctl-api',
        op: 'change',
        componentType: 'helm_chart',
        fields: [
          { key: 'chart_name', op: 'change', diff: "'ctl-api-v1' -> 'ctl-api-v2'" },
          { key: 'namespace', op: 'change', diff: "'default' -> 'app'" },
        ],
        files: [
          {
            name: './values/prod.yaml',
            op: 'change',
            before: 'replicas: 1\nimage:\n  tag: v1\nresources:\n  limits:\n    cpu: 500m\n',
            after: 'replicas: 3\nimage:\n  tag: v2\nresources:\n  limits:\n    cpu: 1000m\n',
          },
          {
            name: './values/feature-flags.yaml',
            op: 'add',
            before: '',
            after: 'featureFlags:\n  beta: true\n  newDashboard: true\n',
          },
        ],
      },
    ],
    fields: [],
    files: [],
  },
  {
    name: 'Actions',
    sectionKey: 'actions',
    grouped: true,
    additions: 1,
    removals: 0,
    changed: 1,
    entities: [
      {
        name: 'run-migrations',
        op: 'add',
        fields: [
          { key: 'timeout', op: 'add', diff: "'' -> '300s'" },
          { key: 'role', op: 'add', diff: "'' -> 'admin'" },
        ],
        files: [],
      },
      {
        name: 'healthcheck',
        op: 'change',
        fields: [{ key: 'role', op: 'change', diff: "'admin' -> 'operator'" }],
        files: [
          {
            name: 'inline_contents',
            op: 'change',
            before: '#!/bin/bash\ncurl -f http://localhost:8080/health\n',
            after: '#!/bin/bash\ncurl -f http://localhost:8080/healthz\nexit 0\n',
          },
        ],
      },
    ],
    fields: [],
    files: [],
  },
  {
    name: 'Runner',
    sectionKey: 'runner',
    grouped: false,
    additions: 0,
    removals: 0,
    changed: 2,
    entities: [],
    fields: [
      { key: 'runner_type', op: 'change', diff: "'standard' -> 'gpu'" },
      { key: 'init_script', op: 'change', diff: "'setup.sh' -> 'setup-gpu.sh'" },
    ],
    files: [],
    content: {
      op: 'change',
      before: 'runner_type = "standard"\nhelm_driver = "secret"\ninit_script = "setup.sh"\n',
      after: 'runner_type = "gpu"\nhelm_driver = "secret"\ninit_script = "setup-gpu.sh"\n',
    },
  },
]

export const Default = () => (
  <AppConfigDiff
    sections={mockSections}
    summary={{ added: 2, removed: 1, changed: 3 }}
  />
)

export const NoChanges = () => (
  <AppConfigDiff sections={[]} summary={null} />
)

export const Loading = () => (
  <AppConfigDiff sections={[]} summary={null} isLoading />
)

export const ComponentsOnly = () => (
  <AppConfigDiff
    sections={[mockSections[0]]}
    summary={{ added: 1, removed: 1, changed: 1 }}
  />
)

export const FieldsOnly = () => (
  <AppConfigDiff
    sections={[mockSections[2]]}
    summary={{ added: 0, removed: 0, changed: 2 }}
  />
)

export const AllSections = () => {
  const allSections: DiffSectionData[] = [
    mockSections[0],
    mockSections[1],
    {
      name: 'Install inputs',
      sectionKey: 'inputs',
      grouped: true,
      additions: 2,
      removals: 0,
      changed: 0,
      entities: [
        {
          name: 'cluster_name',
          op: 'add',
          fields: [
            { key: 'type', op: 'add', diff: "'' -> 'string'" },
            { key: 'required', op: 'add', diff: "'false' -> 'true'" },
          ],
          files: [],
        },
        {
          name: 'region',
          op: 'add',
          fields: [{ key: 'default', op: 'add', diff: "'' -> 'us-west-2'" }],
          files: [],
        },
      ],
      fields: [],
      files: [],
      content: {
        op: 'add',
        before: '',
        after:
          '[[input]]\nname = "cluster_name"\ntype = "string"\nrequired = true\n\n[[input]]\nname = "region"\ndefault = "us-west-2"\n',
      },
    },
    {
      name: 'Secrets',
      sectionKey: 'secrets',
      grouped: true,
      additions: 1,
      removals: 0,
      changed: 0,
      entities: [
        {
          name: 'DATABASE_URL',
          op: 'add',
          fields: [{ key: 'required', op: 'add', diff: "'' -> 'true'" }],
          files: [],
        },
      ],
      fields: [],
      files: [],
      content: {
        op: 'add',
        before: '',
        after: '[[secret]]\nname = "DATABASE_URL"\nrequired = true\n',
      },
    },
    mockSections[2],
    {
      name: 'Stack',
      sectionKey: 'stack',
      grouped: false,
      additions: 0,
      removals: 0,
      changed: 1,
      entities: [],
      fields: [{ key: 'type', op: 'change', diff: "'eks' -> 'eks-v2'" }],
      files: [],
      content: {
        op: 'change',
        before: 'type = "eks"\nname = "platform"\n',
        after: 'type = "eks-v2"\nname = "platform"\n',
      },
    },
    {
      name: 'Sandbox',
      sectionKey: 'sandbox',
      grouped: false,
      additions: 0,
      removals: 0,
      changed: 1,
      entities: [],
      fields: [
        { key: 'terraform_version', op: 'change', diff: "'1.5.0' -> '1.6.0'" },
      ],
      files: [],
      content: {
        op: 'change',
        before: 'terraform_version = "1.5.0"\n\n[public_repo]\nbranch = "main"\n',
        after: 'terraform_version = "1.6.0"\n\n[public_repo]\nbranch = "main"\n',
      },
    },
    {
      name: 'Permissions',
      sectionKey: 'permissions',
      grouped: false,
      additions: 1,
      removals: 0,
      changed: 0,
      entities: [],
      fields: [
        { key: 'provision', op: 'add', diff: "'' -> 'arn:aws:iam::role/deploy'" },
      ],
      files: [],
      content: {
        op: 'add',
        before: '',
        after: '[provision_role]\narn = "arn:aws:iam::role/deploy"\n',
      },
    },
  ]

  return (
    <AppConfigDiff
      sections={allSections}
      summary={{ added: 6, removed: 1, changed: 5 }}
    />
  )
}
