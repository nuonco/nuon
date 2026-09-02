import type { TComponentBuild } from '@/types'
import { BranchRunBuilds } from './BranchRunBuilds'

export default {
  title: 'Branches/BranchRunBuilds',
}

const build = (overrides: Partial<TComponentBuild>): TComponentBuild =>
  ({
    id: 'bld7vv2juioe2cgvr7a4m494n0',
    component_id: 'cmp123',
    component_name: 'ec2',
    component_config_connection: { type: 'terraform_module' },
    status_v2: { status: 'active' },
    ...overrides,
  }) as TComponentBuild

export const SingleBuild = () => (
  <div className="max-w-3xl">
    <BranchRunBuilds builds={[build({})]} orgId="org1" appId="app1" />
  </div>
)

export const MultipleBuilds = () => (
  <div className="max-w-3xl">
    <BranchRunBuilds
      builds={[
        build({
          id: 'bld1',
          component_name: 'ec2',
          component_config_connection: { type: 'terraform_module' },
          status_v2: { status: 'active' },
        }),
        build({
          id: 'bld2',
          component_name: 'api',
          component_config_connection: { type: 'docker_build' },
          status_v2: { status: 'building' },
        }),
        build({
          id: 'bld3',
          component_name: 'ingress',
          component_config_connection: { type: 'helm_chart' },
          status_v2: { status: 'error' },
        }),
      ]}
      orgId="org1"
      appId="app1"
    />
  </div>
)

export const UnknownType = () => (
  <div className="max-w-3xl">
    <BranchRunBuilds
      builds={[
        build({
          id: 'bld9',
          component_name: 'legacy-thing',
          component_config_connection: { type: 'unknown' },
          status_v2: undefined,
        }),
      ]}
      orgId="org1"
      appId="app1"
    />
  </div>
)

export const Empty = () => (
  <div className="max-w-3xl">
    <BranchRunBuilds builds={[]} orgId="org1" appId="app1" />
  </div>
)

export const WithSandboxAndTypes = () => (
  <div className="max-w-3xl">
    <BranchRunBuilds
      builds={[
        build({
          id: 'bld1',
          component_name: 'ec2',
          component_config_connection: { type: 'terraform_module' },
          status_v2: { status: 'active' },
        }),
        build({
          id: 'bld2',
          component_name: 'api',
          component_config_connection: { type: 'docker_build' },
          status_v2: { status: 'building' },
        }),
        build({
          id: 'bld3',
          component_name: 'ingress',
          component_config_connection: { type: 'helm_chart' },
          status_v2: { status: 'error' },
        }),
      ]}
      sandboxBuild={{
        id: 'asb1',
        status: 'active',
        status_v2: { status: 'active' },
      }}
      orgId="org1"
      appId="app1"
    />
  </div>
)
