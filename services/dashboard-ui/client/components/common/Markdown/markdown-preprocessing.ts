export function preprocessContent(content: string): string {
  const lines = content.split('\n')
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
