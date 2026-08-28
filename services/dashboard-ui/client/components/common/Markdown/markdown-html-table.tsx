import {
  Children,
  isValidElement,
  type ReactElement,
  type ReactNode,
} from 'react'
import { DateTime } from 'luxon'
import type {
  ExtractedTable,
  TableCell,
  TableSearchConfig,
} from './markdown-table'

function nodeText(node: ReactNode): string {
  if (node == null || typeof node === 'boolean') return ''
  if (typeof node === 'string' || typeof node === 'number') return String(node)
  if (Array.isArray(node)) return node.map(nodeText).join('')
  if (isValidElement(node)) return nodeText((node.props as any)?.children)
  return ''
}

function tagName(el: ReactElement): string | undefined {
  if (typeof el.type === 'string') return el.type
  return (el.props as any)?.node?.tagName
}

function collectRows(node: ReactNode, out: ReactElement[]): void {
  Children.forEach(node, (child) => {
    if (!isValidElement(child)) return
    if (tagName(child) === 'tr') {
      out.push(child)
      return
    }
    const children = (child.props as any)?.children
    if (children) collectRows(children, out)
  })
}

function rowCells(tr: ReactElement): ReactElement[] {
  const cells: ReactElement[] = []
  Children.forEach((tr.props as any)?.children, (child) => {
    if (isValidElement(child)) {
      const tag = tagName(child)
      if (tag === 'th' || tag === 'td') cells.push(child)
    }
  })
  return cells
}

function span(cell: ReactElement, key: 'colSpan' | 'rowSpan'): number {
  const raw = (cell.props as any)?.[key]
  const n = Number(raw)
  return Number.isFinite(n) ? n : 1
}

function cellSortValue(node: ReactNode): number | undefined {
  let result: number | undefined
  const visit = (n: ReactNode) => {
    if (result !== undefined) return
    if (Array.isArray(n)) {
      n.forEach(visit)
      return
    }
    if (!isValidElement(n)) return
    const props = n.props as any
    const seconds = props?.seconds
    if (
      seconds != null &&
      String(seconds).trim() !== '' &&
      !Number.isNaN(Number(seconds))
    ) {
      const dt = DateTime.fromSeconds(Number(seconds))
      if (dt.isValid) {
        result = dt.toMillis()
        return
      }
    }
    if (typeof props?.time === 'string' && props.time.trim() !== '') {
      const dt = DateTime.fromISO(props.time)
      if (dt.isValid) {
        result = dt.toMillis()
        return
      }
    }
    if (props?.children) visit(props.children)
  }
  visit(node)
  return result
}

function toCell(cell: ReactElement): TableCell {
  const content = (cell.props as any)?.children
  return {
    text: nodeText(content).trim(),
    content,
    sortValue: cellSortValue(content),
  }
}

export function htmlTableToExtracted(
  children: ReactNode,
  search: TableSearchConfig
): ExtractedTable | null {
  const rows: ReactElement[] = []
  collectRows(children, rows)

  const cellRows = rows.map(rowCells).filter((cells) => cells.length > 0)
  if (cellRows.length === 0) return null

  let headerIdx = cellRows.findIndex((cells) =>
    cells.every((c) => tagName(c) === 'th')
  )
  if (headerIdx === -1) headerIdx = 0

  const headerCells = cellRows[headerIdx]
  const colCount = headerCells.length
  if (colCount === 0) return null

  const bodyCellRows = cellRows.filter((_, i) => i !== headerIdx)

  for (const cells of cellRows) {
    for (const cell of cells) {
      if (span(cell, 'colSpan') > 1 || span(cell, 'rowSpan') > 1) return null
    }
  }
  if (bodyCellRows.some((cells) => cells.length !== colCount)) return null

  return {
    headers: headerCells.map(toCell),
    rows: bodyCellRows.map((cells) => cells.map(toCell)),
    align: headerCells.map(() => null),
    search,
  }
}
