const EMBEDDED_CONTENT_KEYS = new Set([
  'content',
  'contents',
  'inline_contents',
  'var_file',
])

const decodeEscapedContent = (value: string, quote: string) =>
  value
    .replace(/\\r\\n/g, '\n')
    .replace(/\\n/g, '\n')
    .replace(/\\r/g, '\n')
    .replace(/\\t/g, '  ')
    .replace(new RegExp(`\\\\${quote}`, 'g'), quote)

export const formatEmbeddedTomlContent = (source: string) =>
  source
    .split('\n')
    .map((line) => {
      const match = line.match(
        /^(\s*)([A-Za-z0-9_-]+)(\s*=\s*)(["'])(.*)\4\s*$/
      )
      if (!match) return line

      const [, indent, key, assignment, quote, value] = match
      if (!EMBEDDED_CONTENT_KEYS.has(key) || !/\\[nrt]/.test(value)) {
        return line
      }

      const decoded = decodeEscapedContent(value, quote)
      const fence = quote.repeat(3)
      return `${indent}${key}${assignment}${fence}\n${decoded}\n${indent}${fence}`
    })
    .join('\n')
