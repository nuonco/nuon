import { ModalStory } from '@/components/__stories__/helpers'
import { ComponentDependencyGraphModal } from './ComponentDependencyGraphContainer'
import type { TAppConfig } from '@/types'

export default {
  title: 'Components/ComponentDependencyGraphModal',
}

const makeConfig = (
  connections: {
    id: string
    name: string
    type: string
    depIds?: string[]
  }[],
): TAppConfig => ({
  component_config_connections: connections.map((c) => ({
    component_id: c.id,
    component_name: c.name,
    type: c.type as TAppConfig['component_config_connections'][0]['type'],
    component_dependency_ids: c.depIds ?? [],
  })),
})

const basePath = '/org-1/apps/app-1/components'

// Typical web app: VPC → EKS cluster → Helm services → frontend
const webAppConfig = makeConfig([
  { id: 'vpc', name: 'vpc', type: 'terraform_module' },
  { id: 'eks', name: 'eks_cluster', type: 'terraform_module', depIds: ['vpc'] },
  { id: 'rds', name: 'postgres', type: 'terraform_module', depIds: ['vpc'] },
  { id: 'api', name: 'api_server', type: 'helm_chart', depIds: ['eks', 'rds'] },
  { id: 'worker', name: 'background_worker', type: 'helm_chart', depIds: ['eks', 'rds'] },
  { id: 'frontend', name: 'frontend', type: 'docker_build', depIds: ['api'] },
  { id: 'monitoring', name: 'observability', type: 'helm_chart', depIds: ['eks'] },
])

export const WebAppFromApiServer = () => (
  <ModalStory label="API server (transitive deps + dependents)">
    <ComponentDependencyGraphModal
      componentId="api"
      componentName="api_server"
      componentType="helm_chart"
      appConfig={webAppConfig}
      basePath={basePath}
    />
  </ModalStory>
)

export const WebAppFromVpc = () => (
  <ModalStory label="VPC (root — all downstream)">
    <ComponentDependencyGraphModal
      componentId="vpc"
      componentName="vpc"
      componentType="terraform_module"
      appConfig={webAppConfig}
      basePath={basePath}
    />
  </ModalStory>
)

export const WebAppFromFrontend = () => (
  <ModalStory label="Frontend (leaf — all upstream)">
    <ComponentDependencyGraphModal
      componentId="frontend"
      componentName="frontend"
      componentType="docker_build"
      appConfig={webAppConfig}
      basePath={basePath}
    />
  </ModalStory>
)

export const WebAppFromEksCluster = () => (
  <ModalStory label="EKS cluster (mid-graph, both directions)">
    <ComponentDependencyGraphModal
      componentId="eks"
      componentName="eks_cluster"
      componentType="terraform_module"
      appConfig={webAppConfig}
      basePath={basePath}
    />
  </ModalStory>
)

// Cross-dependent edges: ALB → kubelogstream, coder → kubelogstream
const crossDepConfig = makeConfig([
  { id: 'cert', name: 'certificate', type: 'terraform_module' },
  { id: 'rds', name: 'rds_cluster_coder', type: 'terraform_module' },
  { id: 'coder', name: 'coder', type: 'helm_chart', depIds: ['cert', 'rds'] },
  { id: 'obs', name: 'observability', type: 'helm_chart', depIds: ['coder'] },
  { id: 'alb', name: 'application_load_balancer', type: 'helm_chart', depIds: ['coder'] },
  { id: 'kls', name: 'kubelogstream', type: 'helm_chart', depIds: ['coder', 'alb'] },
])

export const CrossDependentEdges = () => (
  <ModalStory label="Coder (dependents with cross-edges)">
    <ComponentDependencyGraphModal
      componentId="coder"
      componentName="coder"
      componentType="helm_chart"
      appConfig={crossDepConfig}
      basePath={basePath}
    />
  </ModalStory>
)

// Deep chain: A → B → C → D → E
const deepChainConfig = makeConfig([
  { id: 'a', name: 'network', type: 'terraform_module' },
  { id: 'b', name: 'dns_zone', type: 'terraform_module', depIds: ['a'] },
  { id: 'c', name: 'certificate', type: 'terraform_module', depIds: ['b'] },
  { id: 'd', name: 'load_balancer', type: 'terraform_module', depIds: ['c'] },
  { id: 'e', name: 'ingress', type: 'helm_chart', depIds: ['d'] },
])

export const DeepChainFromMiddle = () => (
  <ModalStory label="Certificate (deep chain, middle node)">
    <ComponentDependencyGraphModal
      componentId="c"
      componentName="certificate"
      componentType="terraform_module"
      appConfig={deepChainConfig}
      basePath={basePath}
    />
  </ModalStory>
)

// Simple: single dependency
const simpleConfig = makeConfig([
  { id: 'infra', name: 'base_infrastructure', type: 'terraform_module' },
  { id: 'svc', name: 'service', type: 'helm_chart', depIds: ['infra'] },
])

export const SimpleOneDependency = () => (
  <ModalStory label="Simple one-dependency">
    <ComponentDependencyGraphModal
      componentId="svc"
      componentName="service"
      componentType="helm_chart"
      appConfig={simpleConfig}
      basePath={basePath}
    />
  </ModalStory>
)
