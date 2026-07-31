export default {
  title: 'Components/ComponentDependencyTable',
}

import { ComponentDependencyTable } from './ComponentDependencyTable'
import type { GraphNode, GraphEdge } from './ComponentDependencyGraph'

const nodes: GraphNode[] = [
  { id: 'cmp-api', name: 'api', type: 'docker_build', role: 'current' },
  { id: 'cmp-db', name: 'database', type: 'terraform_module', role: 'dependency' },
  { id: 'cmp-cache', name: 'cache', type: 'helm_chart', role: 'dependency' },
  { id: 'cmp-worker', name: 'worker', type: 'docker_build', role: 'dependent' },
  { id: 'cmp-web', name: 'web', type: 'docker_build', role: 'dependent' },
]

const edges: GraphEdge[] = [
  { sourceId: 'cmp-db', targetId: 'cmp-api' },
  { sourceId: 'cmp-cache', targetId: 'cmp-api' },
  { sourceId: 'cmp-api', targetId: 'cmp-worker' },
  { sourceId: 'cmp-api', targetId: 'cmp-web' },
]

export const Default = () => (
  <ComponentDependencyTable
    nodes={nodes}
    edges={edges}
    currentId="cmp-api"
    basePath="/org-1/apps/app-1/components"
  />
)

export const SingleDependency = () => (
  <ComponentDependencyTable
    nodes={nodes.slice(0, 2)}
    edges={edges.slice(0, 1)}
    currentId="cmp-api"
    basePath="/org-1/apps/app-1/components"
  />
)

export const LongList = () => {
  const manyNodes: GraphNode[] = [
    { id: 'cmp-root', name: 'root', type: 'terraform_module', role: 'current' },
    ...Array.from({ length: 12 }, (_, i) => ({
      id: `cmp-dep-${i}`,
      name: `service-${i}`,
      type: 'docker_build' as GraphNode['type'],
      role: 'dependent' as const,
    })),
  ]
  const manyEdges: GraphEdge[] = manyNodes
    .slice(1)
    .map((n) => ({ sourceId: 'cmp-root', targetId: n.id }))
  return (
    <ComponentDependencyTable
      nodes={manyNodes}
      edges={manyEdges}
      currentId="cmp-root"
      basePath="/org-1/apps/app-1/components"
    />
  )
}
