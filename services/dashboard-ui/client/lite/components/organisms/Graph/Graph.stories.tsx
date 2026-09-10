import { useMemo, type ComponentType, type ReactNode } from 'react'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { Link } from '../../atoms/Link'
import { Text } from '../../atoms/Text'
import { GraphNodeCard } from '../../molecules/GraphNodeCard'
import {
  Graph,
  GRAPH_NODE_LIMIT,
  type IGraphEdge,
  type IGraphNode,
  type INodeRenderProps,
} from './Graph'

export default {
  title: 'lite/organisms/Graph',
}

type TFixtureData = {
  label: string
  detail?: string
  status?: string
}

const ResourceNode = ({
  href,
  selected,
  data,
}: INodeRenderProps) => {
  const fixture = data as TFixtureData
  return (
    <GraphNodeCard href={href} selected={selected} status={fixture.status}>
      <Text weight="medium">{fixture.label}</Text>
      {fixture.detail ? (
        <Text variant="caption" color="tertiary">
          {fixture.detail}
        </Text>
      ) : null}
    </GraphNodeCard>
  )
}

const ContainerNode = ({ data, selected }: INodeRenderProps) => {
  const fixture = data as TFixtureData
  return (
    <div
      className={
        selected
          ? 'h-full w-full rounded-xl border bg-card-bg/70 graph-node-selected'
          : 'h-full w-full rounded-xl border bg-card-bg/70'
      }
    >
      <Text variant="label" color="secondary" className="px-3 py-2">
        {fixture.label}
      </Text>
    </div>
  )
}

const NODE_TYPES: Record<string, ComponentType<INodeRenderProps>> = {
  resource: ResourceNode,
  container: ContainerNode,
}

const resource = (
  id: string,
  label: string,
  detail: string,
  href?: string,
  status?: string
): IGraphNode => ({
  id,
  kind: 'resource',
  width: 208,
  height: 80,
  href,
  data: { label, detail, status },
})

const container = (id: string, label: string): IGraphNode => ({
  id,
  kind: 'container',
  width: 240,
  height: 80,
  data: { label },
})

const FLAT_NODES: IGraphNode[] = [
  resource('api', 'payments-api', 'Helm chart', '?panel=component:api', 'healthy'),
  resource('worker', 'payments-worker', 'Kubernetes manifest', '?panel=component:worker', 'deploying'),
  resource('db', 'payments-db', 'Terraform module', '?panel=component:db', 'healthy'),
]

const FLAT_EDGES: IGraphEdge[] = [
  { id: 'api-db', source: 'api', target: 'db' },
  { id: 'worker-db', source: 'worker', target: 'db' },
]

const NESTED_NODES: IGraphNode[] = [
  container('cluster', 'production cluster'),
  {
    ...resource('api', 'payments-api', 'Helm chart', '?panel=component:api', 'healthy'),
    parentId: 'cluster',
  },
  {
    ...resource(
      'worker',
      'payments-worker',
      'Kubernetes manifest',
      '?panel=component:worker',
      'deploying'
    ),
    parentId: 'cluster',
  },
  resource('db', 'payments-db', 'Terraform module', '?panel=component:db', 'healthy'),
]

const NESTED_EDGES: IGraphEdge[] = [
  { id: 'api-worker', source: 'api', target: 'worker' },
]

const CROSS_NODES: IGraphNode[] = NESTED_NODES

const CROSS_EDGES: IGraphEdge[] = [
  { id: 'api-worker', source: 'api', target: 'worker' },
  { id: 'api-db', source: 'api', target: 'db' },
  { id: 'worker-db', source: 'worker', target: 'db' },
]

const Frame = ({ children }: { children: ReactNode }) => (
  <div className="p-8">{children}</div>
)

export const Overview = () => (
  <ComponentDocs
    name="Graph"
    tier="organism"
    summary="A read-only map of nodes and edges, laid out hierarchically, whose nodes open panels."
    use={[
      'Show how parts of a system connect, including containers with children and edges that leave a container.',
      'Pass a resolved graph. Graph renders what it is handed and does not fetch or know a resource.',
    ]}
    avoid={[
      'Do not use a graph when the user is comparing records field by field. That is Table.',
      'Do not use a graph for a linear sequence of two or three steps. That is a list or a timeline.',
      'Do not hand-roll a canvas. Compose this organism.',
    ]}
    rules={[
      'The caller owns node sizes. Graph does not measure the DOM.',
      'Containment is parentId. Graph sorts parents before children because React Flow requires it.',
      'Layout is async and keyed on node ids, parent ids, and edge endpoints. A status-only update repaints without relaying out.',
      'A node with href is a Link. Graph never calls openPanel.',
      'Graph owns loading, fetching, empty, failed, and too-large. Empty and failed share the emptyState slot with different copy.',
      'Nodes are not draggable and edges are not editable. There is no minimap and no attribution.',
    ]}
    props={[
      {
        name: 'nodes',
        type: 'IGraphNode[]',
        description: 'Nodes to lay out. kind selects the renderer from nodeTypes.',
      },
      {
        name: 'edges',
        type: 'IGraphEdge[]',
        description: 'Directed edges. kind is a free string for maps to tag a style.',
      },
      {
        name: 'nodeTypes',
        type: 'Record<string, ComponentType<INodeRenderProps>>',
        description: 'Renderer for each node kind. Resource nodes compose GraphNodeCard.',
      },
      {
        name: 'direction',
        type: "'right' | 'down'",
        default: "'right'",
        description: 'Layout direction. The only public ELK option.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description:
          'First load with no data. Also shown until the first layout resolves.',
      },
      {
        name: 'fetching',
        type: 'boolean',
        default: 'false',
        description: 'Revalidating with a graph on screen. Must not blank or relayout.',
      },
      {
        name: 'emptyState',
        type: 'ReactNode',
        default: "'No nodes yet'",
        description:
          'Centered empty and failed copy. Failed uses this slot with failure wording.',
      },
      {
        name: 'height',
        type: "number | 'auto'",
        default: '480',
        description: 'Canvas height in pixels, or auto to grow with the layout.',
      },
    ]}
    sections={[
      {
        heading: 'Containment',
        body: (
          <div className="flex flex-col gap-3">
            <Text color="secondary">
              parentId places a node inside another. Graph sizes the parent from
              its children. Edges may leave a container; that is why layout is ELK
              rather than a flat ranker.
            </Text>
            <Graph
              nodes={CROSS_NODES}
              edges={CROSS_EDGES}
              nodeTypes={NODE_TYPES}
              height={360}
            />
          </div>
        ),
      },
    ]}
  />
)

export const Flat = () => (
  <Frame>
    <Graph nodes={FLAT_NODES} edges={FLAT_EDGES} nodeTypes={NODE_TYPES} />
  </Frame>
)

export const Nested = () => (
  <Frame>
    <Graph nodes={NESTED_NODES} edges={NESTED_EDGES} nodeTypes={NODE_TYPES} />
  </Frame>
)

export const CrossContainer = () => (
  <Frame>
    <Graph nodes={CROSS_NODES} edges={CROSS_EDGES} nodeTypes={NODE_TYPES} />
  </Frame>
)

export const Selection = () => (
  <div className="flex flex-col gap-3 p-8">
    <Text variant="caption" color="tertiary">
      Click a node. Edges that do not touch it dim; the node keeps its selected
      outline.
    </Text>
    <Graph nodes={CROSS_NODES} edges={CROSS_EDGES} nodeTypes={NODE_TYPES} />
  </div>
)

export const Loading = () => (
  <Frame>
    <Graph nodes={[]} edges={[]} nodeTypes={NODE_TYPES} loading />
  </Frame>
)

export const Fetching = () => (
  <Frame>
    <Graph
      nodes={FLAT_NODES}
      edges={FLAT_EDGES}
      nodeTypes={NODE_TYPES}
      fetching
    />
  </Frame>
)

export const Empty = () => (
  <Frame>
    <Graph nodes={[]} edges={[]} nodeTypes={NODE_TYPES} />
  </Frame>
)

export const Failed = () => (
  <Frame>
    <Graph
      nodes={[]}
      edges={[]}
      nodeTypes={NODE_TYPES}
      emptyState="Graph failed to load"
    />
  </Frame>
)

export const TooLarge = () => {
  const nodes = useMemo(
    () =>
      Array.from({ length: GRAPH_NODE_LIMIT + 1 }, (_, index) =>
        resource(`n${index}`, `node-${index}`, 'Fixture')
      ),
    []
  )

  return (
    <Frame>
      <Graph
        nodes={nodes}
        edges={[]}
        nodeTypes={NODE_TYPES}
        emptyState={<Link href="/org-example/installs">View as a list</Link>}
      />
    </Frame>
  )
}

export const Directions = () => (
  <div className="flex flex-col gap-8 p-8">
    <div className="flex flex-col gap-2">
      <Text variant="label" color="tertiary">
        Right
      </Text>
      <Graph
        nodes={FLAT_NODES}
        edges={FLAT_EDGES}
        nodeTypes={NODE_TYPES}
        direction="right"
        height={320}
      />
    </div>
    <div className="flex flex-col gap-2">
      <Text variant="label" color="tertiary">
        Down
      </Text>
      <Graph
        nodes={FLAT_NODES}
        edges={FLAT_EDGES}
        nodeTypes={NODE_TYPES}
        direction="down"
        height={360}
      />
    </div>
  </div>
)

export const Keyboard = () => (
  <div className="flex flex-col gap-3 p-8">
    <Text variant="caption" color="tertiary">
      Tab through the linked nodes. Fit to view is in the corner if a node sits
      off-canvas.
    </Text>
    <Graph nodes={FLAT_NODES} edges={FLAT_EDGES} nodeTypes={NODE_TYPES} />
  </div>
)
