import { describe, expect, test } from 'bun:test'
import { parseDotGraph } from './parse-dot'

const dotFixture = `
digraph {
  "cmp-db" [ label="database", type="terraform_module", color="blue" ];
  "cmp-api" [ label="api", type="docker_build", color="red" ];
  "cmp-db" -> "cmp-api" [ color="red" ];
  "cmp-api" -> "cmp-worker" [ ];
}
`

describe('parseDotGraph', () => {
  test('parses nodes with attributes', () => {
    const { nodes } = parseDotGraph(dotFixture)
    const db = nodes.find((n) => n.id === 'cmp-db')
    expect(db).toEqual({
      id: 'cmp-db',
      label: 'database',
      type: 'terraform_module',
      changed: true,
    })
    const api = nodes.find((n) => n.id === 'cmp-api')
    expect(api?.changed).toBe(false)
    expect(api?.label).toBe('api')
  })

  test('parses edges with and without attributes', () => {
    const { edges } = parseDotGraph(dotFixture)
    expect(edges).toEqual([
      { source: 'cmp-db', target: 'cmp-api', color: 'red' },
      { source: 'cmp-api', target: 'cmp-worker', color: '' },
    ])
  })

  test('creates fallback nodes for ids that only appear in edges', () => {
    const { nodes } = parseDotGraph(dotFixture)
    const worker = nodes.find((n) => n.id === 'cmp-worker')
    expect(worker).toEqual({
      id: 'cmp-worker',
      label: 'cmp-worker',
      type: '',
      changed: false,
    })
  })

  test('falls back to name attribute then id for labels', () => {
    const { nodes } = parseDotGraph(
      '"n1" [ name="from-name" ];\n"n2" [ type="helm_chart" ];'
    )
    expect(nodes.find((n) => n.id === 'n1')?.label).toBe('from-name')
    expect(nodes.find((n) => n.id === 'n2')?.label).toBe('n2')
  })

  test('returns empty for empty input', () => {
    expect(parseDotGraph('')).toEqual({ nodes: [], edges: [] })
  })
})
