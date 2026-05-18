import { ComponentDependencyGraph, type GraphNode, type GraphEdge } from './ComponentDependencyGraph'

export default {
  title: 'Components/ComponentDependencyGraph',
}

const basePath = '/org-1/apps/app-1/components'

const webAppNodes: GraphNode[] = [
  { id: 'vpc', name: 'vpc', type: 'terraform_module', role: 'dependency' },
  { id: 'rds', name: 'postgres', type: 'terraform_module', role: 'dependency' },
  { id: 'eks', name: 'eks_cluster', type: 'terraform_module', role: 'dependency' },
  { id: 'api', name: 'api_server', type: 'helm_chart', role: 'current' },
  { id: 'frontend', name: 'frontend', type: 'docker_build', role: 'dependent' },
  { id: 'worker', name: 'background_worker', type: 'helm_chart', role: 'dependent' },
]

const webAppEdges: GraphEdge[] = [
  { sourceId: 'vpc', targetId: 'eks' },
  { sourceId: 'vpc', targetId: 'rds' },
  { sourceId: 'eks', targetId: 'api' },
  { sourceId: 'rds', targetId: 'api' },
  { sourceId: 'api', targetId: 'frontend' },
  { sourceId: 'api', targetId: 'worker' },
  { sourceId: 'eks', targetId: 'worker' },
]

export const FullTransitiveGraph = () => (
  <div style={{ width: 700, height: 500 }}>
    <ComponentDependencyGraph
      nodes={webAppNodes}
      edges={webAppEdges}
      currentId="api"
      basePath={basePath}
    />
  </div>
)

export const DependenciesOnly = () => (
  <div style={{ width: 600, height: 400 }}>
    <ComponentDependencyGraph
      nodes={[
        { id: 'vpc', name: 'vpc', type: 'terraform_module', role: 'dependency' },
        { id: 'api', name: 'api_server', type: 'helm_chart', role: 'dependency' },
        { id: 'frontend', name: 'frontend', type: 'docker_build', role: 'current' },
      ]}
      edges={[
        { sourceId: 'vpc', targetId: 'api' },
        { sourceId: 'api', targetId: 'frontend' },
      ]}
      currentId="frontend"
      basePath={basePath}
    />
  </div>
)

export const DependentsOnly = () => (
  <div style={{ width: 600, height: 400 }}>
    <ComponentDependencyGraph
      nodes={[
        { id: 'vpc', name: 'vpc', type: 'terraform_module', role: 'current' },
        { id: 'eks', name: 'eks_cluster', type: 'terraform_module', role: 'dependent' },
        { id: 'rds', name: 'database', type: 'terraform_module', role: 'dependent' },
        { id: 'api', name: 'api_server', type: 'helm_chart', role: 'dependent' },
      ]}
      edges={[
        { sourceId: 'vpc', targetId: 'eks' },
        { sourceId: 'vpc', targetId: 'rds' },
        { sourceId: 'eks', targetId: 'api' },
        { sourceId: 'rds', targetId: 'api' },
      ]}
      currentId="vpc"
      basePath={basePath}
    />
  </div>
)

export const SingleNode = () => (
  <div style={{ width: 600, height: 400 }}>
    <ComponentDependencyGraph
      nodes={[{ id: 'standalone', name: 'standalone', type: 'helm_chart', role: 'current' }]}
      edges={[]}
      currentId="standalone"
      basePath={basePath}
    />
  </div>
)

export const CrossLevelEdges = () => (
  <div style={{ width: 700, height: 500 }}>
    <ComponentDependencyGraph
      nodes={[
        { id: 'cert', name: 'certificate', type: 'terraform_module', role: 'dependency' },
        { id: 'rds', name: 'rds_cluster', type: 'terraform_module', role: 'dependency' },
        { id: 'coder', name: 'coder', type: 'helm_chart', role: 'current' },
        { id: 'obs', name: 'observability', type: 'helm_chart', role: 'dependent' },
        { id: 'alb', name: 'application_load_balancer', type: 'helm_chart', role: 'dependent' },
        { id: 'kls', name: 'kubelogstream', type: 'helm_chart', role: 'dependent' },
      ]}
      edges={[
        { sourceId: 'cert', targetId: 'coder' },
        { sourceId: 'rds', targetId: 'coder' },
        { sourceId: 'coder', targetId: 'obs' },
        { sourceId: 'coder', targetId: 'alb' },
        { sourceId: 'coder', targetId: 'kls' },
        { sourceId: 'alb', targetId: 'kls' },
      ]}
      currentId="coder"
      basePath={basePath}
    />
  </div>
)
