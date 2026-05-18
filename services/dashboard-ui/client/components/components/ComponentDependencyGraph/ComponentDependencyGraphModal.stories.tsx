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
  <ModalStory label="API server graph">
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
  <ModalStory label="VPC graph (root dependency)">
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
  <ModalStory label="Frontend graph (leaf dependent)">
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
  <ModalStory label="EKS cluster graph (many dependents)">
    <ComponentDependencyGraphModal
      componentId="eks"
      componentName="eks_cluster"
      componentType="terraform_module"
      appConfig={webAppConfig}
      basePath={basePath}
    />
  </ModalStory>
)

// Infrastructure-heavy app: networking → compute → services
const infraConfig = makeConfig([
  { id: 'network', name: 'network', type: 'terraform_module' },
  { id: 'dns', name: 'dns_zone', type: 'terraform_module', depIds: ['network'] },
  { id: 'cert', name: 'certificate', type: 'terraform_module', depIds: ['dns'] },
  { id: 'alb', name: 'load_balancer', type: 'terraform_module', depIds: ['network', 'cert'] },
  { id: 'cluster', name: 'eks_cluster', type: 'terraform_module', depIds: ['network'] },
  { id: 'ingress', name: 'ingress_controller', type: 'helm_chart', depIds: ['cluster', 'alb'] },
  { id: 'app', name: 'application', type: 'helm_chart', depIds: ['cluster', 'ingress'] },
  { id: 'jobs', name: 'cron_jobs', type: 'kubernetes_manifest', depIds: ['cluster'] },
])

export const InfraFromLoadBalancer = () => (
  <ModalStory label="Load balancer graph">
    <ComponentDependencyGraphModal
      componentId="alb"
      componentName="load_balancer"
      componentType="terraform_module"
      appConfig={infraConfig}
      basePath={basePath}
    />
  </ModalStory>
)

export const InfraFromIngressController = () => (
  <ModalStory label="Ingress controller graph">
    <ComponentDependencyGraphModal
      componentId="ingress"
      componentName="ingress_controller"
      componentType="helm_chart"
      appConfig={infraConfig}
      basePath={basePath}
    />
  </ModalStory>
)

// Simple app: single dependency
const simpleConfig = makeConfig([
  { id: 'infra', name: 'base_infrastructure', type: 'terraform_module' },
  { id: 'svc', name: 'service', type: 'helm_chart', depIds: ['infra'] },
])

export const SimpleOneDependency = () => (
  <ModalStory label="Simple one-dependency graph">
    <ComponentDependencyGraphModal
      componentId="svc"
      componentName="service"
      componentType="helm_chart"
      appConfig={simpleConfig}
      basePath={basePath}
    />
  </ModalStory>
)

// Component with no deps or dependents (shouldn't normally open, but useful for edge case)
const isolatedConfig = makeConfig([
  { id: 'standalone', name: 'standalone_job', type: 'job' },
  { id: 'other', name: 'other_service', type: 'helm_chart' },
])

export const IsolatedComponent = () => (
  <ModalStory label="Isolated component (no connections)">
    <ComponentDependencyGraphModal
      componentId="standalone"
      componentName="standalone_job"
      componentType="job"
      appConfig={isolatedConfig}
      basePath={basePath}
    />
  </ModalStory>
)
