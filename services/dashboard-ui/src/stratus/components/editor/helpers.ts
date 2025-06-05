export function splitYamlDiff(diff: string) {
  const originalLines = []
  const modifiedLines = []

  diff.split('\n').forEach((line) => {
    if (line.startsWith('-')) {
      originalLines.push(line.slice(1).trimStart())
    } else if (line.startsWith('+')) {
      modifiedLines.push(line.slice(1).trimStart())
    } else if (line.trim() !== '') {

      originalLines.push(line)
      modifiedLines.push(line)
    }
  })

  return {
    original: originalLines.join('\n'),
    modified: modifiedLines.join('\n'),
  }
}
