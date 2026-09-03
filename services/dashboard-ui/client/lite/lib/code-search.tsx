import type { KeyboardEvent } from 'react'
import type { FileDiffMetadata, SelectionSide } from '@pierre/diffs/react'
import { Icon } from '../components/atoms/Icon'
import { Kbd } from '../components/atoms/Kbd'
import { Text } from '../components/atoms/Text'

export interface IDiffMatch {
  lineNumber: number
  side: SelectionSide
}

export const lineMatches = (value: string, query: string) => {
  const needle = query.trim().toLowerCase()
  if (!needle) return []

  return value
    .split('\n')
    .flatMap((line, index) =>
      line.toLowerCase().includes(needle) ? [index + 1] : []
    )
}

export const diffMatches = (
  fileDiff: FileDiffMetadata,
  query: string
): IDiffMatch[] => {
  const needle = query.trim().toLowerCase()
  if (!needle) return []

  const { additionLines, deletionLines } = fileDiff
  const found: IDiffMatch[] = []

  const collect = (
    lines: string[],
    start: number,
    count: number,
    side: SelectionSide
  ) => {
    for (let index = start; index < start + count; index++) {
      if (lines[index]?.toLowerCase().includes(needle)) {
        found.push({ lineNumber: index + 1, side })
      }
    }
  }

  let cursor = 0

  for (const hunk of fileDiff.hunks) {
    collect(
      additionLines,
      cursor,
      Math.max(hunk.additionLineIndex - cursor, 0),
      'additions'
    )
    cursor = Math.max(cursor, hunk.additionLineIndex)

    for (const block of hunk.hunkContent) {
      if (block.type === 'context') {
        collect(
          additionLines,
          block.additionLineIndex,
          block.lines,
          'additions'
        )
        cursor = Math.max(cursor, block.additionLineIndex + block.lines)
        continue
      }

      collect(
        deletionLines,
        block.deletionLineIndex,
        block.deletions,
        'deletions'
      )
      collect(
        additionLines,
        block.additionLineIndex,
        block.additions,
        'additions'
      )
      cursor = Math.max(cursor, block.additionLineIndex + block.additions)
    }
  }

  collect(additionLines, cursor, additionLines.length - cursor, 'additions')

  return found
}

export const matchNavKeyDown =
  (matchIndex: number, goTo: (index: number) => void) =>
  (event: KeyboardEvent<HTMLInputElement>) => {
    const back =
      event.key === 'ArrowUp' || (event.key === 'Enter' && event.shiftKey)
    const forward = event.key === 'ArrowDown' || event.key === 'Enter'
    if (!back && !forward) return

    event.preventDefault()
    goTo(back ? matchIndex - 1 : matchIndex + 1)
  }

const tooltip = (label: string, key: 'ArrowUpIcon' | 'ArrowDownIcon') => (
  <span className="flex items-center gap-1.5">
    <Text variant="caption">{label}</Text>
    <Kbd>
      <Icon variant={key} size={10} />
    </Kbd>
  </span>
)

export const MATCH_NAV_TOOLTIP = {
  previous: tooltip('Previous match', 'ArrowUpIcon'),
  next: tooltip('Next match', 'ArrowDownIcon'),
}
