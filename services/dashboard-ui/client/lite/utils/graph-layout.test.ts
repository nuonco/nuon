import { describe, expect, test } from 'bun:test'
import {
  graphLayoutSignature,
  layoutGraph,
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

describe('graph layout', () => {
  test('places children inside their parent bounds', async () => {
    const positioned = await layoutGraph({
      nodes: [node('group'), node('first', 'group'), node('second', 'group')],
      edges: [{ id: 'inside', source: 'first', target: 'second' }],
    })
    const parent = positioned.find(({ id }) => id === 'group')!
    const children = positioned.filter(({ parentId }) => parentId === parent.id)

    for (const child of children) {
      expect(child.position.x).toBeGreaterThanOrEqual(0)
      expect(child.position.y).toBeGreaterThanOrEqual(0)
      expect(child.position.x + child.width).toBeLessThanOrEqual(parent.width)
      expect(child.position.y + child.height).toBeLessThanOrEqual(parent.height)
    }
  })

  test('keeps both endpoints of a cross-container edge', async () => {
    const edges: IGraphLayoutEdge[] = [
      { id: 'cross', source: 'inside', target: 'outside' },
    ]
    const positioned = await layoutGraph({
      nodes: [node('group'), node('inside', 'group'), node('outside')],
      edges,
    })

    expect(positioned.map(({ id }) => id).sort()).toEqual([
      'group',
      'inside',
      'outside',
    ])
    expect(edges).toEqual([
      { id: 'cross', source: 'inside', target: 'outside' },
    ])
  })

  test('emits parents before their children', async () => {
    const positioned = await layoutGraph({
      nodes: [node('grandchild', 'child'), node('child', 'parent'), node('parent')],
      edges: [],
    })
    const ids = positioned.map(({ id }) => id)

    expect(ids.indexOf('parent')).toBeLessThan(ids.indexOf('child'))
    expect(ids.indexOf('child')).toBeLessThan(ids.indexOf('grandchild'))
  })

  test('moves a node with a missing parent to the root', async () => {
    const positioned = await layoutGraph({
      nodes: [node('orphan', 'missing')],
      edges: [],
    })

    expect(positioned[0]?.id).toBe('orphan')
    expect(positioned[0]?.parentId).toBeUndefined()
  })

  test('terminates when the edges contain a cycle', async () => {
    const positioned = await layoutGraph({
      nodes: [node('first'), node('second')],
      edges: [
        { id: 'forward', source: 'first', target: 'second' },
        { id: 'backward', source: 'second', target: 'first' },
      ],
    })

    expect(positioned.map(({ id }) => id).sort()).toEqual(['first', 'second'])
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
