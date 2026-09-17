import { describe, expect, test } from 'bun:test'
import {
  graphLayoutSignature,
  layoutGraph,
  orthogonalEdgePath,
  type IGraphLayoutEdge,
  type IGraphLayoutNode,
} from './graph-layout'

const node = (
  id: string,
  parentId?: string,
  data: Record<string, unknown> = {}
): IGraphLayoutNode & { data: Record<string, unknown> } => ({
  id,
  parentId,
  width: 120,
  height: 56,
  data,
})

const overlaps = (
  first: { x: number; y: number; width: number; height: number },
  second: { x: number; y: number; width: number; height: number }
) =>
  first.x < second.x + second.width &&
  second.x < first.x + first.width &&
  first.y < second.y + second.height &&
  second.y < first.y + first.height

describe('graph layout', () => {
  test('places children inside their parent bounds', async () => {
    const { nodes } = await layoutGraph({
      nodes: [node('group'), node('first', 'group'), node('second', 'group')],
      edges: [{ id: 'inside', source: 'first', target: 'second' }],
    })
    const parent = nodes.find(({ id }) => id === 'group')!
    const children = nodes.filter(({ parentId }) => parentId === parent.id)

    expect(children).toHaveLength(2)
    for (const child of children) {
      expect(child.position.x).toBeGreaterThanOrEqual(0)
      expect(child.position.y).toBeGreaterThanOrEqual(0)
      expect(child.position.x + child.width).toBeLessThanOrEqual(parent.width)
      expect(child.position.y + child.height).toBeLessThanOrEqual(parent.height)
    }
  })

  test('does not overlap sibling nodes', async () => {
    const { nodes } = await layoutGraph({
      nodes: [node('group'), node('first', 'group'), node('second', 'group')],
      edges: [{ id: 'inside', source: 'first', target: 'second' }],
    })
    const [first, second] = nodes.filter(({ parentId }) => parentId === 'group')

    expect(
      overlaps(
        { ...first!.position, width: first!.width, height: first!.height },
        { ...second!.position, width: second!.width, height: second!.height }
      )
    ).toBe(false)
  })

  test('keeps both endpoints of a cross-container edge', async () => {
    const edges: IGraphLayoutEdge[] = [
      { id: 'cross', source: 'inside', target: 'outside' },
    ]
    const { nodes } = await layoutGraph({
      nodes: [node('group'), node('inside', 'group'), node('outside')],
      edges,
    })

    expect(nodes.map(({ id }) => id).sort()).toEqual([
      'group',
      'inside',
      'outside',
    ])
    expect(edges).toEqual([
      { id: 'cross', source: 'inside', target: 'outside' },
    ])
  })

  test('routes every edge with absolute bend points', async () => {
    const { nodes, edges } = await layoutGraph({
      nodes: [node('group'), node('inside', 'group'), node('outside')],
      edges: [
        { id: 'cross', source: 'inside', target: 'outside' },
        { id: 'back', source: 'outside', target: 'inside' },
      ],
    })
    const inside = nodes.find(({ id }) => id === 'inside')!
    const group = nodes.find(({ id }) => id === 'group')!
    const insideLeft = group.position.x + inside.position.x

    for (const edge of edges) {
      expect(edge.points.length).toBeGreaterThan(1)
    }

    const cross = edges.find(({ id }) => id === 'cross')!
    expect(cross.points.at(0)!.x).toBeGreaterThanOrEqual(insideLeft)
  })

  test('emits parents before their children', async () => {
    const { nodes } = await layoutGraph({
      nodes: [
        node('grandchild', 'child'),
        node('child', 'parent'),
        node('parent'),
      ],
      edges: [],
    })
    const ids = nodes.map(({ id }) => id)

    expect(ids.indexOf('parent')).toBeLessThan(ids.indexOf('child'))
    expect(ids.indexOf('child')).toBeLessThan(ids.indexOf('grandchild'))
  })

  test('moves a node with a missing parent to the root', async () => {
    const { nodes } = await layoutGraph({
      nodes: [node('orphan', 'missing')],
      edges: [],
    })

    expect(nodes[0]?.id).toBe('orphan')
    expect(nodes[0]?.parentId).toBeUndefined()
  })

  test('terminates when the edges contain a cycle', async () => {
    const { nodes } = await layoutGraph({
      nodes: [node('first'), node('second')],
      edges: [
        { id: 'forward', source: 'first', target: 'second' },
        { id: 'backward', source: 'second', target: 'first' },
      ],
    })

    expect(nodes.map(({ id }) => id).sort()).toEqual(['first', 'second'])
  })
})

describe('orthogonal edge path', () => {
  test('returns an empty path without points', () => {
    expect(orthogonalEdgePath([])).toBe('')
  })

  test('draws a straight run between two points', () => {
    expect(orthogonalEdgePath([{ x: 0, y: 10 }, { x: 50, y: 10 }])).toBe(
      'M 0 10 L 50 10'
    )
  })

  test('rounds a corner without overshooting a short segment', () => {
    const path = orthogonalEdgePath(
      [
        { x: 0, y: 0 },
        { x: 4, y: 0 },
        { x: 4, y: 40 },
      ],
      8
    )

    expect(path).toContain('Q 4 0')
    expect(path).toMatch(/^M 0 0 /)
    expect(path.endsWith('L 4 40')).toBe(true)
  })
})

describe('graph layout signature', () => {
  const nodes = [node('first'), node('second')]
  const edges = [{ id: 'edge', source: 'first', target: 'second' }]

  test('changes when graph structure changes', () => {
    const original = graphLayoutSignature({ nodes, edges })
    const parentChanged = graphLayoutSignature({
      nodes: [node('first'), node('second', 'first')],
      edges,
    })
    const endpointChanged = graphLayoutSignature({
      nodes,
      edges: [{ id: 'edge', source: 'second', target: 'first' }],
    })

    expect(parentChanged).not.toBe(original)
    expect(endpointChanged).not.toBe(original)
  })

  test('does not change for status-only updates or input order', () => {
    const original = graphLayoutSignature({ nodes, edges })
    const updated = graphLayoutSignature({
      nodes: [
        node('second', undefined, { status: 'failed' }),
        node('first', undefined, { status: 'running' }),
      ],
      edges: [...edges],
    })

    expect(updated).toBe(original)
  })
})
