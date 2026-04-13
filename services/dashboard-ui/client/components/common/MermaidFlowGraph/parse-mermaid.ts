export type ParsedNode = {
  id: string
  label: string
  nodeShape: 'rect' | 'rounded' | 'diamond' | 'cylinder' | 'circle' | 'asymmetric' | 'parallelogram' | 'subroutine'
  style?: { fill?: string; stroke?: string; color?: string }
}

export type ParsedEdge = {
  source: string
  target: string
  label?: string
  type: 'solid' | 'dashed' | 'thick'
  hasArrow: boolean
}

export type ParsedSubgraph = {
  id: string
  label: string
  children: string[]
}

export type ParseResult = {
  direction: string
  nodes: ParsedNode[]
  edges: ParsedEdge[]
  subgraphs: ParsedSubgraph[]
}

function parseNodeDeclaration(token: string): { id: string; label: string; nodeShape: ParsedNode['nodeShape'] } | null {
  const patterns: Array<{ re: RegExp; nodeShape: ParsedNode['nodeShape'] }> = [
    { re: /^(\w+)\[\[(.+?)\]\]$/, nodeShape: 'subroutine' },
    { re: /^(\w+)\[\((.+?)\)\]$/, nodeShape: 'cylinder' },
    { re: /^(\w+)\[\/(.+?)\/\]$/, nodeShape: 'parallelogram' },
    { re: /^(\w+)\(\((.+?)\)\)$/, nodeShape: 'circle' },
    { re: /^(\w+)\((.+?)\)$/, nodeShape: 'rounded' },
    { re: /^(\w+)\{(.+?)\}$/, nodeShape: 'diamond' },
    { re: /^(\w+)>(.+?)\]$/, nodeShape: 'asymmetric' },
    { re: /^(\w+)\[(.+?)\]$/, nodeShape: 'rect' },
  ]

  for (const { re, nodeShape } of patterns) {
    const m = token.match(re)
    if (m) {
      return { id: m[1], label: m[2].replace(/<br\s*\/?>/gi, '\n').replace(/"/g, ''), nodeShape }
    }
  }

  return null
}

function parseEdge(line: string): { source: string; target: string; label?: string; type: ParsedEdge['type']; hasArrow: boolean; sourceDecl?: string; targetDecl?: string } | null {
  const edgeRe = /^(.+?)\s+(={3,}>|={2}>|-{3,}>?|--?>|-\.->|\.->|={2,}>)\s*(?:\|([^|]*)\|)?\s*(.+)$/
  const m = line.match(edgeRe)
  if (!m) return null

  const [, rawSource, arrow, label, rawTarget] = m

  let type: ParsedEdge['type'] = 'solid'
  let hasArrow = true

  if (arrow.includes('.')) {
    type = 'dashed'
  } else if (arrow.startsWith('=')) {
    type = 'thick'
  }

  if (arrow === '---' || arrow === '----' || /^-{3,}$/.test(arrow)) {
    hasArrow = false
  }

  const sourceId = rawSource.trim().match(/^(\w+)/)?.[1] || rawSource.trim()
  const targetId = rawTarget.trim().match(/^(\w+)/)?.[1] || rawTarget.trim()

  return {
    source: sourceId,
    target: targetId,
    label: label?.trim(),
    type,
    hasArrow,
    sourceDecl: rawSource.trim(),
    targetDecl: rawTarget.trim(),
  }
}

function parseStyleDirective(line: string): { nodeId: string; styles: Record<string, string> } | null {
  const m = line.match(/^style\s+(\w+)\s+(.+)$/)
  if (!m) return null

  const [, nodeId, styleStr] = m
  const styles: Record<string, string> = {}

  for (const part of styleStr.split(',')) {
    const [key, val] = part.split(':').map((s) => s.trim())
    if (key && val) styles[key] = val
  }

  return { nodeId, styles }
}

function collectAllDescendantNodes(
  sg: ParsedSubgraph,
  allSubgraphs: ParsedSubgraph[],
  nodeMap: Map<string, ParsedNode>,
): string[] {
  const result: string[] = []
  for (const childId of sg.children) {
    if (nodeMap.has(childId)) {
      result.push(childId)
    } else {
      const childSg = allSubgraphs.find((s) => s.id === childId)
      if (childSg) {
        result.push(...collectAllDescendantNodes(childSg, allSubgraphs, nodeMap))
      }
    }
  }
  return result
}

export function parseMermaidFlowchart(code: string): ParseResult {
  const lines = code.split('\n').map((l) => l.trim()).filter(Boolean)

  let direction = 'TB'
  const nodes = new Map<string, ParsedNode>()
  const edges: ParsedEdge[] = []
  const subgraphs: ParsedSubgraph[] = []
  const styleMap = new Map<string, Record<string, string>>()

  const headerMatch = lines[0]?.match(/^(?:graph|flowchart)\s+(TD|TB|LR|RL|BT)\s*$/i)
  if (headerMatch) {
    direction = headerMatch[1].toUpperCase()
    if (direction === 'TD') direction = 'TB'
  }

  const subgraphStack: ParsedSubgraph[] = []

  function isSubgraphId(id: string): boolean {
    return subgraphs.some((sg) => sg.id === id)
  }

  function ensureNode(id: string) {
    if (isSubgraphId(id)) return
    if (!nodes.has(id)) {
      nodes.set(id, { id, label: id, nodeShape: 'rect' })
    }
    if (subgraphStack.length > 0) {
      const current = subgraphStack[subgraphStack.length - 1]
      if (!current.children.includes(id)) {
        current.children.push(id)
      }
    }
  }

  function tryRegisterNode(token: string): string | null {
    const parsed = parseNodeDeclaration(token)
    if (parsed) {
      if (!nodes.has(parsed.id)) {
        nodes.set(parsed.id, { id: parsed.id, label: parsed.label, nodeShape: parsed.nodeShape })
      } else {
        const existing = nodes.get(parsed.id)!
        if (existing.label === existing.id) {
          existing.label = parsed.label
          existing.nodeShape = parsed.nodeShape
        }
      }
      if (subgraphStack.length > 0) {
        const current = subgraphStack[subgraphStack.length - 1]
        if (!current.children.includes(parsed.id)) {
          current.children.push(parsed.id)
        }
      }
      return parsed.id
    }
    return null
  }

  const startIdx = headerMatch ? 1 : 0

  for (let i = startIdx; i < lines.length; i++) {
    const line = lines[i]

    if (/^%%/.test(line)) continue

    const subgraphMatch = line.match(/^subgraph\s+(\w+)\s*\["([^"]+)"\]\s*$/) ||
      line.match(/^subgraph\s+(\w+)\s*\[([^\]]+)\]\s*$/) ||
      line.match(/^subgraph\s+(\w+)\s*$/)
    if (subgraphMatch) {
      const sgId = subgraphMatch[1]
      const sgLabel = subgraphMatch[2] || sgId
      const sg: ParsedSubgraph = {
        id: sgId,
        label: sgLabel.trim(),
        children: [],
      }
      if (subgraphStack.length > 0) {
        const parent = subgraphStack[subgraphStack.length - 1]
        if (!parent.children.includes(sgId)) {
          parent.children.push(sgId)
        }
      }
      subgraphs.push(sg)
      subgraphStack.push(sg)
      continue
    }

    if (!subgraphMatch && /^subgraph\s+/.test(line)) {
      const fallbackLabel = line.replace(/^subgraph\s+/, '').trim()
      const sgId = `subgraph_${subgraphs.length}`
      const sg: ParsedSubgraph = {
        id: sgId,
        label: fallbackLabel,
        children: [],
      }
      if (subgraphStack.length > 0) {
        const parent = subgraphStack[subgraphStack.length - 1]
        if (!parent.children.includes(sgId)) {
          parent.children.push(sgId)
        }
      }
      subgraphs.push(sg)
      subgraphStack.push(sg)
      continue
    }

    if (line === 'end') {
      subgraphStack.pop()
      continue
    }

    const styleResult = parseStyleDirective(line)
    if (styleResult) {
      styleMap.set(styleResult.nodeId, styleResult.styles)
      continue
    }

    if (/^classDef\s/.test(line) || /^class\s/.test(line) || /^click\s/.test(line) || /^linkStyle\s/.test(line)) {
      continue
    }

    if (/~~~/.test(line)) continue

    const edgeResult = parseEdge(line)
    if (edgeResult) {
      if (edgeResult.sourceDecl) tryRegisterNode(edgeResult.sourceDecl)
      if (edgeResult.targetDecl) tryRegisterNode(edgeResult.targetDecl)
      ensureNode(edgeResult.source)
      ensureNode(edgeResult.target)
      edges.push({
        source: edgeResult.source,
        target: edgeResult.target,
        label: edgeResult.label,
        type: edgeResult.type,
        hasArrow: edgeResult.hasArrow,
      })
      continue
    }

    tryRegisterNode(line) || ensureNode(line.match(/^(\w+)$/)?.[1] || '')
  }

  for (const [nodeId, styles] of styleMap) {
    const node = nodes.get(nodeId)
    if (node) {
      node.style = {
        fill: styles.fill,
        stroke: styles.stroke,
        color: styles.color,
      }
    }
  }

  nodes.delete('')

  const subgraphIdMap = new Map(subgraphs.map((sg) => [sg.id, sg]))

  // Remove any phantom nodes that share an ID with a subgraph
  for (const sgId of subgraphIdMap.keys()) {
    nodes.delete(sgId)
  }

  // Rewire edges that reference subgraph IDs to the first real child node
  for (let i = edges.length - 1; i >= 0; i--) {
    const edge = edges[i]
    let removed = false
    for (const ref of [edge.source, edge.target] as const) {
      if (subgraphIdMap.has(ref)) {
        const sg = subgraphIdMap.get(ref)!
        const allChildren = collectAllDescendantNodes(sg, subgraphs, nodes)
        const firstChild = allChildren[0]
        if (firstChild) {
          if (ref === edge.source) edge.source = firstChild
          else edge.target = firstChild
        } else {
          edges.splice(i, 1)
          removed = true
          break
        }
      }
    }
    if (!removed && !nodes.has(edge.source)) edges.splice(i, 1)
    else if (!removed && !nodes.has(edge.target)) edges.splice(i, 1)
  }

  return {
    direction,
    nodes: Array.from(nodes.values()),
    edges,
    subgraphs,
  }
}
