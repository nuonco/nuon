const ALERT_HEADER = /^>\s*\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*$/

function preprocessAlerts(content: string): string {
  const lines = content.split('\n')
  const result: string[] = []
  let i = 0

  while (i < lines.length) {
    const headerMatch = ALERT_HEADER.exec(lines[i])
    if (!headerMatch) {
      result.push(lines[i])
      i++
      continue
    }

    const bodyLines: string[] = []
    i++

    while (i < lines.length && /^>\s?/.test(lines[i])) {
      bodyLines.push(lines[i].replace(/^>\s?/, ''))
      i++
    }

    const body = bodyLines.join('\n').trim()
    result.push(`<nuon-alert type="${headerMatch[1].toLowerCase()}">${body}</nuon-alert>`)
    result.push('')
  }

  return result.join('\n')
}

export function preprocessContent(content: string): string {
  const alerted = preprocessAlerts(content)
  const lines = alerted.split('\n')
  const result: string[] = []
  let htmlDepth = 0

  for (const line of lines) {
    const opens = (line.match(/<(?:div|table|thead|tbody|tr|ul|ol|section)\b/gi) || []).length
    const closes = (line.match(/<\/(?:div|table|thead|tbody|tr|ul|ol|section)\b/gi) || []).length
    htmlDepth += opens - closes

    if (htmlDepth > 0 && line.trim() === '') {
      continue
    }

    result.push(line)

    if (htmlDepth < 0) htmlDepth = 0
  }

  return result.join('\n')
}
