export default {
  title: 'Approvals/PlanDiffs/InstallGroupDiff',
}

import { InstallGroupDiff } from './InstallGroupDiff'
import type { InstallDiffEntry } from './InstallGroupDiff'
import type { DiffEntityEntry, DiffSectionData } from '../app-config/AppConfigDiff'

const componentsSection = (entities: DiffEntityEntry[]): DiffSectionData => ({
  name: 'Components',
  sectionKey: 'components',
  grouped: true,
  additions: entities.filter((e) => e.op === 'add').length,
  removals: entities.filter((e) => e.op === 'remove').length,
  changed: entities.filter((e) => e.op === 'change').length,
  entities,
  fields: [],
})

const infrastructureSection = (sandbox: boolean, stack: boolean): DiffSectionData => {
  const entities: DiffEntityEntry[] = [
    ...(sandbox ? [{ name: 'Sandbox', op: 'change' as const, fields: [] }] : []),
    ...(stack ? [{ name: 'Stack', op: 'change' as const, fields: [] }] : []),
  ]
  return {
    name: 'Infrastructure',
    sectionKey: 'infrastructure',
    grouped: true,
    additions: 0,
    removals: 0,
    changed: entities.length,
    entities,
    fields: [],
  }
}

const changedComponents: InstallDiffEntry = {
  installId: 'inst-abc123',
  installName: 'production-us-west-2',
  installLabels: { env: 'production', region: 'us-west-2' },
  status: 'pending',
  summary: { added: 0, removed: 0, changed: 3 },
  sections: [
    componentsSection([
      { name: 'certificate', op: 'change', componentType: 'terraform_module', fields: [] },
      { name: 'coder', op: 'change', componentType: 'helm_chart', fields: [] },
      { name: 'observability', op: 'change', componentType: 'helm_chart', fields: [] },
    ]),
  ],
}

const withInfrastructure: InstallDiffEntry = {
  installId: 'inst-def456',
  installName: 'jm-test',
  installLabels: { env: 'uat' },
  status: 'pending',
  sandboxChanged: true,
  stackChanged: true,
  summary: { added: 0, removed: 0, changed: 2 },
  sections: [
    componentsSection([
      { name: 'rds_subnet', op: 'change', componentType: 'terraform_module', fields: [] },
      { name: 'kubelogstream', op: 'change', componentType: 'helm_chart', fields: [] },
    ]),
    infrastructureSection(true, true),
  ],
}

const mixedOps: InstallDiffEntry = {
  installId: 'inst-ghi789',
  installName: 'staging-us-east-1',
  installLabels: { env: 'staging', region: 'us-east-1' },
  status: 'pending',
  sandboxChanged: true,
  stackChanged: false,
  summary: { added: 1, removed: 1, changed: 1 },
  sections: [
    componentsSection([
      { name: 'redis', op: 'add', componentType: 'helm_chart', fields: [] },
      { name: 'application_load_balancer', op: 'change', componentType: 'helm_chart', fields: [] },
      { name: 'legacy_worker', op: 'remove', componentType: 'docker_build', fields: [] },
    ]),
    infrastructureSection(true, false),
  ],
}

const noChanges: InstallDiffEntry = {
  installId: 'inst-jkl012',
  installName: 'dev-local',
  status: 'pending',
  summary: { added: 0, removed: 0, changed: 0 },
  sections: [],
}

export const Default = () => (
  <InstallGroupDiff groupName="UAT" installs={[changedComponents, withInfrastructure, mixedOps, noChanges]} />
)

export const SingleInstall = () => (
  <InstallGroupDiff groupName="production" installs={[changedComponents]} />
)

export const WithSandboxAndStack = () => (
  <InstallGroupDiff groupName="UAT" installs={[withInfrastructure]} />
)

export const MixedOps = () => (
  <InstallGroupDiff groupName="staging" installs={[mixedOps]} />
)

export const NoChanges = () => (
  <InstallGroupDiff groupName="dev" installs={[noChanges]} />
)

export const Empty = () => <InstallGroupDiff groupName="canary" installs={[]} />

export const Loading = () => <InstallGroupDiff groupName="UAT" installs={[]} isLoading />
