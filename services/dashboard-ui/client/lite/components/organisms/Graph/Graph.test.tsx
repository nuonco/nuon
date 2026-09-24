import type { ComponentType } from 'react'
import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { Text } from '../../atoms/Text'
import { GraphNodeCard } from '../../molecules/GraphNodeCard'
import {
  Graph,
  GRAPH_NODE_LIMIT,
  type IGraphEdge,
  type IGraphNode,
  type INodeRenderProps,
} from './Graph'

type TNodeData = {
  label: string
}

const NodeRenderer = ({
  data,
  href,
  selected,
}: INodeRenderProps) => {
  const node = data as TNodeData

  return (
    <GraphNodeCard href={href} selected={selected}>
      <Text>{node.label}</Text>
    </GraphNodeCard>
  )
}

const NODE_TYPES: Record<string, ComponentType<INodeRenderProps>> = {
  resource: NodeRenderer,
}

const node = (id: string, label: string, href?: string): IGraphNode => ({
  id,
  kind: 'resource',
  width: 160,
  height: 64,
  data: { label },
  href,
})

const NODES = [
  node('linked', 'Linked node', '/panels/linked'),
  node('static', 'Static node'),
]

const EDGES: IGraphEdge[] = [
  { id: 'linked-static', source: 'linked', target: 'static' },
]

const Example = ({
  nodes = NODES,
  edges = EDGES,
  loading,
  emptyState,
}: {
  nodes?: IGraphNode[]
  edges?: IGraphEdge[]
  loading?: boolean
  emptyState?: string
}) => (
  <MemoryRouter>
    <Graph
      nodes={nodes}
      edges={edges}
      nodeTypes={NODE_TYPES}
      loading={loading}
      emptyState={emptyState}
    />
  </MemoryRouter>
)

afterEach(cleanup)

describe('Graph', () => {
  test('renders linked and non-interactive nodes with distinct semantics', async () => {
    render(<Example />)

    expect(
      await screen.findByRole('link', { name: 'Linked node' })
    ).toHaveAttribute('href', '/panels/linked')
    expect(screen.getByText('Static node').closest('a')).toBeNull()
    expect(screen.getAllByRole('link')).toHaveLength(1)
  })

  test('renders distinct loading, empty, and failed content', async () => {
    const { rerender } = render(<Example nodes={[]} edges={[]} loading />)

    expect(screen.getByRole('status', { name: 'Loading graph' })).toBeTruthy()
    expect(screen.queryByText('No nodes yet')).toBeNull()

    rerender(<Example nodes={[]} edges={[]} />)
    expect(screen.getByText('No nodes yet')).toBeTruthy()
    expect(
      screen.queryByRole('status', { name: 'Loading graph' })
    ).toBeNull()

    rerender(
      <Example
        nodes={[]}
        edges={[]}
        emptyState="Graph failed to load"
      />
    )
    expect(screen.getByText('Graph failed to load')).toBeTruthy()
    expect(screen.queryByText('No nodes yet')).toBeNull()
  })

  test('renders the too-large message instead of nodes', () => {
    const nodes = Array.from({ length: GRAPH_NODE_LIMIT + 1 }, (_, index) =>
      node(`node-${index}`, `Node ${index}`)
    )

    render(<Example nodes={nodes} edges={[]} />)

    expect(
      screen.getByText(
        `This graph is too large to display. Graphs with more than ${GRAPH_NODE_LIMIT} nodes are shown as a list instead.`
      )
    ).toBeTruthy()
    expect(screen.queryByText('Node 0')).toBeNull()
  })

  test('gives every graph control an accessible name', async () => {
    render(<Example />)

    await waitFor(() =>
      expect(
        screen.getByRole('button', { name: 'Zoom in' })
      ).toBeTruthy()
    )
    expect(screen.getByRole('button', { name: 'Zoom out' })).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Fit to view' })).toBeTruthy()
  })
})
