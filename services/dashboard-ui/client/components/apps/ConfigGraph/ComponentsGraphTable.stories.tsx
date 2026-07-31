export default {
  title: 'Apps/ComponentsGraphTable',
}

import { ComponentsGraphTable } from './ComponentsGraphTable'
import type { TDotNode, TDotEdge } from './parse-dot'

const nodes: TDotNode[] = [
  { id: 'cmp-db', label: 'database', type: 'terraform_module', changed: true },
  { id: 'cmp-api', label: 'api', type: 'docker_build', changed: false },
  { id: 'cmp-cache', label: 'cache', type: 'helm_chart', changed: false },
  { id: 'cmp-worker', label: 'worker', type: 'docker_build', changed: true },
]

const edges: TDotEdge[] = [
  { source: 'cmp-db', target: 'cmp-api', color: 'red' },
  { source: 'cmp-cache', target: 'cmp-api', color: 'red' },
  { source: 'cmp-api', target: 'cmp-worker', color: 'red' },
]

export const Default = () => (
  <ComponentsGraphTable nodes={nodes} edges={edges} />
)

export const NoDependencies = () => (
  <ComponentsGraphTable nodes={nodes} edges={[]} />
)

export const Loading = () => (
  <ComponentsGraphTable nodes={[]} edges={[]} isLoading />
)

export const Empty = () => <ComponentsGraphTable nodes={[]} edges={[]} />
