import { useState } from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Select } from '@/components/common/form/Select'
import type { DiffSectionData } from '@/components/branches/AppConfigDiff'
import { BranchOverview } from '@/components/branches/BranchOverview/BranchOverview'
import { AppConfigFilesDiff, ComponentConfigDiff } from './ComponentConfigDiff'

export default {
  title: 'Features / Branches / Component config diff',
}

const valuesBaseBefore = `replicaCount: 1
image:
  repository: ghcr.io/acme/api
  tag: v1.10.1
service:
  port: 80
`

const valuesBaseAfter = `replicaCount: 3
image:
  repository: ghcr.io/acme/api
  tag: v1.10.2
service:
  port: 80
`

const valuesProdBefore = `resources:
  requests:
    cpu: 250m
    memory: 512Mi
ingress:
  enabled: true
  className: nginx
`

const valuesProdAfter = `resources:
  requests:
    cpu: 500m
    memory: 1Gi
ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt
`

const valuesStaging = `resources:
  requests:
    cpu: 250m
    memory: 512Mi
`

const helmConfigChanged: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    additions: 0,
    removals: 0,
    changed: 1,
    grouped: true,
    fields: [],
    entities: [
      {
        name: 'api',
        op: 'change',
        componentType: 'helm_chart',
        fields: [
          {
            key: 'replicaCount',
            op: 'change',
            diff: "'1' -> '3'",
          },
        ],
      },
    ],
  },
]

const helmValuesFiles = [
  {
    path: 'charts/api/values-base.yaml',
    kind: 'helm values' as const,
    change: 'modified' as const,
    before: valuesBaseBefore,
    after: valuesBaseAfter,
  },
  {
    path: 'charts/api/values-prod.yaml',
    kind: 'helm values' as const,
    change: 'modified' as const,
    before: valuesProdBefore,
    after: valuesProdAfter,
  },
  {
    path: 'charts/api/values-staging.yaml',
    kind: 'helm values' as const,
    change: 'unchanged' as const,
    after: valuesStaging,
  },
]

const valuesProdOnlyFiles = [
  {
    path: 'charts/api/values-base.yaml',
    kind: 'helm values' as const,
    change: 'unchanged' as const,
    after: valuesBaseBefore,
  },
  {
    path: 'charts/api/values-prod.yaml',
    kind: 'helm values' as const,
    change: 'modified' as const,
    before: valuesProdBefore,
    after: valuesProdAfter,
  },
]

const tfvarsBefore = `instance_type = "t3.small"
region = "us-west-2"
`

const tfvarsAfter = `instance_type = "t3.medium"
region = "us-west-2"
`

const mainTf = `module "cache" {
  source = "./modules/cache"
}
`

const terraformVarsFiles = [
  {
    path: 'components/cache/main.tf',
    kind: 'terraform source' as const,
    change: 'unchanged' as const,
    after: mainTf,
  },
  {
    path: 'components/cache/terraform.tfvars',
    kind: 'terraform vars' as const,
    change: 'modified' as const,
    before: tfvarsBefore,
    after: tfvarsAfter,
  },
]

const dockerfile = `FROM ghcr.io/acme/api-base:2
COPY bin/api /usr/local/bin/api
ENTRYPOINT ["/usr/local/bin/api"]
`

const legacyManifest = `apiVersion: v1
kind: Service
metadata:
  name: legacy-svc
`

const sandboxMainBefore = `resource "aws_instance" "sandbox" {
  instance_type = "t3.medium"
}
`

const sandboxMainAfter = `resource "aws_instance" "sandbox" {
  instance_type = "t3.large"
}
`

const permissionsBefore = `{
  "statements": [
    { "sid": "DescribeNodegroups", "effect": "Allow", "actions": ["eks:DescribeNodegroup"] },
    { "sid": "ListLoadBalancers", "effect": "Allow", "actions": ["elasticloadbalancing:DescribeLoadBalancers"] }
  ]
}
`

const permissionsAfter = `{
  "statements": [
    { "sid": "DescribeNodegroups", "effect": "Allow", "actions": ["eks:DescribeNodegroup"] },
    { "sid": "ListLoadBalancers", "effect": "Allow", "actions": ["elasticloadbalancing:DescribeLoadBalancers"] },
    { "sid": "ReadSecrets", "effect": "Allow", "actions": ["secretsmanager:GetSecretValue"] }
  ]
}
`

const opaPolicyBefore = `package nuon.api

allow {
  input.method == "GET"
  input.user.roles[_] == "viewer"
}
`

const opaPolicyAfter = `package nuon.api

allow {
  input.method == "GET"
  input.user.roles[_] == "viewer"
}

allow {
  input.method == "POST"
  input.user.roles[_] == "editor"
}
`

const restartScriptBefore = `#!/bin/sh
kubectl rollout restart deployment/api
`

const restartScriptAfter = `#!/bin/sh
kubectl rollout restart deployment/api
kubectl rollout status deployment/api --timeout=120s
`

const fullAppSections: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    additions: 1,
    removals: 1,
    changed: 2,
    grouped: true,
    fields: [],
    entities: [
      {
        name: 'api',
        op: 'change',
        componentType: 'helm_chart',
        fields: [],
        files: [{ name: 'charts/api/values-prod.yaml', op: 'change' }],
      },
      {
        name: 'cache',
        op: 'change',
        componentType: 'terraform_module',
        fields: [],
        files: [{ name: 'components/cache/terraform.tfvars', op: 'change' }],
      },
      {
        name: 'worker',
        op: 'add',
        componentType: 'docker_build',
        fields: [],
        files: [{ name: 'components/worker/Dockerfile', op: 'add' }],
      },
      {
        name: 'legacy',
        op: 'remove',
        componentType: 'kubernetes_manifest',
        fields: [],
        files: [{ name: 'components/legacy/manifest.yaml', op: 'remove' }],
      },
    ],
  },
  {
    name: 'Actions',
    sectionKey: 'actions',
    additions: 0,
    removals: 0,
    changed: 1,
    grouped: true,
    fields: [],
    entities: [
      {
        name: 'restart',
        op: 'change',
        fields: [],
        files: [{ name: 'actions/restart/restart.sh', op: 'change' }],
      },
    ],
  },
  {
    name: 'Runbooks',
    sectionKey: 'runbooks',
    additions: 1,
    removals: 0,
    changed: 0,
    grouped: true,
    fields: [],
    entities: [
      {
        name: 'rotate-logs',
        op: 'add',
        fields: [],
        files: [{ name: 'runbooks/rotate-logs/rotate.sh', op: 'add' }],
      },
    ],
  },
  {
    name: 'Sandbox',
    sectionKey: 'sandbox',
    additions: 0,
    removals: 0,
    changed: 1,
    grouped: false,
    fields: [],
    entities: [],
    files: [{ name: 'sandbox/main.tf', op: 'change' }],
  },
  {
    name: 'Permissions',
    sectionKey: 'permissions',
    additions: 0,
    removals: 0,
    changed: 1,
    grouped: false,
    fields: [],
    entities: [],
    files: [{ name: 'permissions/api.json', op: 'change' }],
  },
  {
    name: 'Policies',
    sectionKey: 'policies',
    additions: 0,
    removals: 0,
    changed: 1,
    grouped: true,
    fields: [],
    entities: [
      {
        name: 'api',
        op: 'change',
        fields: [],
        files: [{ name: 'policies/api.rego', op: 'change' }],
      },
    ],
  },
]

const fullAppFiles = [
  {
    path: 'charts/api/values-base.yaml',
    kind: 'helm values' as const,
    change: 'unchanged' as const,
    after: valuesBaseBefore,
  },
  {
    path: 'charts/api/values-prod.yaml',
    kind: 'helm values' as const,
    change: 'modified' as const,
    before: valuesProdBefore,
    after: valuesProdAfter,
  },
  {
    path: 'components/cache/terraform.tfvars',
    kind: 'terraform vars' as const,
    change: 'modified' as const,
    before: tfvarsBefore,
    after: tfvarsAfter,
  },
  {
    path: 'components/worker/Dockerfile',
    kind: 'dockerfile' as const,
    change: 'added' as const,
    after: dockerfile,
  },
  {
    path: 'components/legacy/manifest.yaml',
    kind: 'kubernetes manifest' as const,
    change: 'removed' as const,
    before: legacyManifest,
  },
  {
    path: 'actions/restart/restart.sh',
    kind: 'action script' as const,
    change: 'modified' as const,
    before: restartScriptBefore,
    after: restartScriptAfter,
  },
  {
    path: 'runbooks/rotate-logs/rotate.sh',
    kind: 'runbook script' as const,
    change: 'added' as const,
    after: restartScriptBefore,
  },
  {
    path: 'sandbox/main.tf',
    kind: 'sandbox terraform' as const,
    change: 'modified' as const,
    before: sandboxMainBefore,
    after: sandboxMainAfter,
  },
  {
    path: 'permissions/api.json',
    kind: 'iam policy' as const,
    change: 'modified' as const,
    before: permissionsBefore,
    after: permissionsAfter,
  },
  {
    path: 'policies/api.rego',
    kind: 'opa policy' as const,
    change: 'modified' as const,
    before: opaPolicyBefore,
    after: opaPolicyAfter,
  },
]

const Callout = ({ children }: { children: string }) => (
  <div className="flex items-start gap-2 rounded-md bg-info-50 px-3 py-2 dark:bg-info-900/30">
    <Icon variant="InfoIcon" size={14} className="mt-0.5 shrink-0" />
    <Text variant="subtext" theme="neutral">
      {children}
    </Text>
  </div>
)

export const FullAppConfig = () => (
  <div className="mx-auto max-w-5xl p-8">
    <Callout>
      App scope: every section of the app config diff, with the referenced
      files of all components collected into one tree.
    </Callout>
    <AppConfigFilesDiff
      previousVersion="#118"
      currentVersion="#119"
      configSections={fullAppSections}
      files={fullAppFiles}
    />
  </div>
)

export const InBranchOverview = () => (
  <BranchOverview
    hasPlan={false}
    rollout={{
      id: 'wf_184',
      href: '#run',
      source: { kind: 'manual' },
      title: 'feat: remove preview_ping action',
      sha: '960a77709002e4b851bed23457a4b79f98ca6422',
      author: 'jane@example.com',
      status: 'success',
    }}
    changes={
      <AppConfigFilesDiff
        title="Template and source changes"
        previousVersion="960a777"
        currentVersion="e4f5a6b"
        configSections={fullAppSections}
        files={fullAppFiles}
        headerAction={
          <Text as="span" variant="subtext" className="text-link">
            View builds
          </Text>
        }
      />
    }
    groups={[]}
    rolloutHref="#rollout"
    groupHref={(id) => `#rollout/groups/${id}`}
  />
)
InBranchOverview.storyName = 'In branch overview'

const branchRefs = [
  { name: 'main', sha: 'a1b2c3d' },
  { name: 'ht/branch-config-diff', sha: 'e4f5a6b' },
  { name: 'feature/auto-scaling', sha: '9c8d7e6' },
  { name: 'release-1.4', sha: '5f4e3d2' },
]

const unchangedFiles = fullAppFiles.map((file) => ({
  ...file,
  change: 'unchanged' as const,
  before: file.before ?? file.after,
}))

export const BranchComparison = () => {
  const [base, setBase] = useState('main')
  const [head, setHead] = useState('ht/branch-config-diff')
  const baseSha = branchRefs.find((b) => b.name === base)?.sha
  const headSha = branchRefs.find((b) => b.name === head)?.sha
  const sameBranch = base === head

  return (
    <div className="mx-auto max-w-5xl p-8">
      <Callout>
        Branch scope: pick a base and a head branch — the diff comes from the
        run comparison (git diff between commit SHAs + app config diff between
        the runs).
      </Callout>
      <div className="mb-4 flex items-end gap-2">
        <div className="w-56">
          <Select
            size="sm"
            aria-label="Base branch"
            options={branchRefs.map((b) => ({ value: b.name, label: b.name }))}
            value={base}
            onChange={setBase}
          />
        </div>
        <Icon variant="ArrowRightIcon" size={16} className="mb-2.5 shrink-0" />
        <div className="w-56">
          <Select
            size="sm"
            aria-label="Compare branch"
            options={branchRefs.map((b) => ({ value: b.name, label: b.name }))}
            value={head}
            onChange={setHead}
          />
        </div>
      </div>
      <AppConfigFilesDiff
        previousVersion={`${base}@${baseSha}`}
        currentVersion={`${head}@${headSha}`}
        configSections={sameBranch ? [] : fullAppSections}
        files={sameBranch ? unchangedFiles : fullAppFiles}
      />
    </div>
  )
}

export const ConfigAndFilesChanged = () => (
  <div className="mx-auto max-w-5xl p-8">
    <ComponentConfigDiff
      componentName="api"
      componentType="helm_chart"
      previousVersion="#118"
      currentVersion="#119"
      configSections={helmConfigChanged}
      files={helmValuesFiles}
    />
  </div>
)

export const ValuesFileOnly = () => (
  <div className="mx-auto max-w-5xl p-8">
    <Callout>
      Only a referenced file changed. The app config diff is empty, but this
      change still deploys — the file tree is what surfaces it.
    </Callout>
    <ComponentConfigDiff
      componentName="api"
      componentType="helm_chart"
      previousVersion="#118"
      currentVersion="#119"
      configSections={[]}
      files={valuesProdOnlyFiles}
    />
  </div>
)

export const TerraformVarsFileOnly = () => (
  <div className="mx-auto max-w-5xl p-8">
    <Callout>
      Same idea for terraform_module: an instance type change in tfvars never
      appears in the app config diff.
    </Callout>
    <ComponentConfigDiff
      componentName="cache"
      componentType="terraform_module"
      previousVersion="#118"
      currentVersion="#119"
      configSections={[]}
      files={terraformVarsFiles}
    />
  </div>
)

export const NoChanges = () => (
  <div className="mx-auto max-w-5xl p-8">
    <ComponentConfigDiff
      componentName="api"
      componentType="helm_chart"
      previousVersion="#118"
      currentVersion="#118"
      configSections={[]}
      files={helmValuesFiles.map((file) => ({ ...file, change: 'unchanged' as const, before: undefined, after: file.after ?? file.before }))}
    />
  </div>
)
