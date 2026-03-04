export function diffLines(
  before: string | object | any[] | undefined | null,
  after: string | object | any[] | undefined | null
): string {
  // Helper to stringify input for diffing
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

  const beforeLines = beforeStr.trim().split('\n')
  const afterLines = afterStr.trim().split('\n')
  const maxLen = Math.max(beforeLines.length, afterLines.length)
  const lines: string[] = []

  for (let i = 0; i < maxLen; i++) {
    const b = beforeLines[i] || ''
    const a = afterLines[i] || ''
    if (b === a) {
      lines.push(`  ${a}`) // unchanged
    } else {
      if (b) lines.push(`- ${b}`) // removed
      if (a) lines.push(`+ ${a}`) // added
    }
  }

  const result = lines.join('\n')

  // Return "No diff to show" if the result is empty or just whitespace
  return result.trim() === '' ? 'No diff to show' : result
}
