export type TDotNode = {
  id: string
  label: string
  type: string
  changed: boolean
}

export type TDotEdge = {
  source: string
  target: string
  color: string
}

const parseAttributes = (attrs: string) => {
  const attributes: Record<string, string> = {}
  const attrRegex = /(\w+)\s*=\s*"([^"]*)"/g
  let attrMatch
  while ((attrMatch = attrRegex.exec(attrs)) !== null) {
    attributes[attrMatch[1]] = attrMatch[2]
  }
  return attributes
}

export function parseDotGraph(dotGraph: string): {
  nodes: TDotNode[]
  edges: TDotEdge[]
} {
  const nodesMap = new Map<string, TDotNode>()
  const edges: TDotEdge[] = []
  const allNodeIds = new Set<string>()

  const nodeWithAttrsRegex = /^\s*"([^"]+)"\s*\[\s*([^\]]+?)\s*\];?\s*$/gm
  let match
  while ((match = nodeWithAttrsRegex.exec(dotGraph)) !== null) {
    const [, id, attrs] = match
    allNodeIds.add(id)
    const attributes = parseAttributes(attrs)
    nodesMap.set(id, {
      id: String(id),
      label: String(attributes.label || attributes.name || id),
      type: String(attributes.type || ''),
      changed: attributes.color === 'blue',
    })
  }

  const edgeRegex =
    /^\s*"([^"]+)"\s*->\s*"([^"]+)"\s*\[\s*([^\]]*)\s*\];?\s*$/gm
  while ((match = edgeRegex.exec(dotGraph)) !== null) {
    const [, source, target, attrs] = match
    allNodeIds.add(source)
    allNodeIds.add(target)
    const attributes = parseAttributes(attrs)
    edges.push({
      source: String(source),
      target: String(target),
      color: attributes.color ?? '',
    })
  }

  allNodeIds.forEach((id) => {
    if (!nodesMap.has(id)) {
      nodesMap.set(id, {
        id: String(id),
        label: String(id),
        type: '',
        changed: false,
      })
    }
  })

  return { nodes: Array.from(nodesMap.values()), edges }
}
