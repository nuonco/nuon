function yamlKeyPrefix(line: string): string | null {
  const match = line.match(/^(\s*[\w./-]+)\s*:/)
  if (match) return match[1]
  const listMatch = line.match(/^(\s*-\s+[\w./-]+)\s*:/)
  if (listMatch) return listMatch[1]
  return null
}

function lcs(a: string[], b: string[]): { type: 'equal' | 'remove' | 'add'; line: string }[] {
  const n = a.length
  const m = b.length

  const dp: number[][] = Array.from({ length: n + 1 }, () => new Array(m + 1).fill(0))
  for (let i = n - 1; i >= 0; i--) {
    for (let j = m - 1; j >= 0; j--) {
      if (a[i] === b[j]) {
        dp[i][j] = dp[i + 1][j + 1] + 1
      } else {
        dp[i][j] = Math.max(dp[i + 1][j], dp[i][j + 1])
      }
    }
  }

  const result: { type: 'equal' | 'remove' | 'add'; line: string }[] = []
  let i = 0
  let j = 0
  while (i < n || j < m) {
    if (i < n && j < m && a[i] === b[j]) {
      result.push({ type: 'equal', line: a[i] })
      i++
      j++
    } else if (i < n && (j >= m || dp[i + 1][j] >= dp[i][j + 1])) {
      result.push({ type: 'remove', line: a[i] })
      i++
    } else {
      result.push({ type: 'add', line: b[j] })
      j++
    }
  }

  return result
}

export function diffLines(
  before: string | object | any[] | undefined | null,
  after: string | object | any[] | undefined | null
): string {
  function toString(val: unknown): string {
    if (typeof val === 'string') return val
    if (val === undefined || val === null) return ''
    if (Array.isArray(val) || typeof val === 'object') {
      try {
        return JSON.stringify(val, null, 2)
      } catch {
        return String(val)
      }
    }
    return String(val)
  }

  const beforeStr = toString(before)
  const afterStr = toString(after)

  const beforeArr = beforeStr.trim().split('\n')
  const afterArr = afterStr.trim().split('\n')

  if (beforeArr.length === 1 && beforeArr[0] === '') beforeArr.length = 0
  if (afterArr.length === 1 && afterArr[0] === '') afterArr.length = 0

  const ops = lcs(beforeArr, afterArr)
  const lines: string[] = []

  let idx = 0
  while (idx < ops.length) {
    if (ops[idx].type === 'equal') {
      lines.push(`  ${ops[idx].line}`)
      idx++
      continue
    }

    const removes: string[] = []
    const adds: string[] = []
    while (idx < ops.length && ops[idx].type === 'remove') {
      removes.push(ops[idx].line)
      idx++
    }
    while (idx < ops.length && ops[idx].type === 'add') {
      adds.push(ops[idx].line)
      idx++
    }

    const pairCount = Math.min(removes.length, adds.length)
    for (let p = 0; p < pairCount; p++) {
      const bLine = removes[p]
      const aLine = adds[p]
      const bKey = yamlKeyPrefix(bLine)
      const aKey = yamlKeyPrefix(aLine)
      if (bKey && aKey && bKey === aKey) {
        const bVal = bLine.slice(bLine.indexOf(':', bKey.length) + 1).trim()
        const aVal = aLine.slice(aLine.indexOf(':', aKey.length) + 1).trim()
        lines.push(
          `~ ${aLine.slice(0, aLine.indexOf(':', aKey.length) + 1)} ${bVal} -> ${aVal}`
        )
      } else {
        lines.push(`- ${bLine}`)
        lines.push(`+ ${aLine}`)
      }
    }
    for (let p = pairCount; p < removes.length; p++) {
      lines.push(`- ${removes[p]}`)
    }
    for (let p = pairCount; p < adds.length; p++) {
      lines.push(`+ ${adds[p]}`)
    }
  }

  const result = lines.join('\n')
  return result.trim() === '' ? 'Diff not available from planner' : result
}
