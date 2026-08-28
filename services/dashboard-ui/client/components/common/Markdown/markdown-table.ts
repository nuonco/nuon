import type { ReactNode } from 'react'

export type TableAlign = 'left' | 'center' | 'right' | null

export interface TableCell {
  text: string
  markdown?: string
  content?: ReactNode
  sortValue?: number
}

export interface TableSearchConfig {
  columns: string[] | null
  placeholder?: string
}

export interface ExtractedTable {
  headers: TableCell[]
  rows: TableCell[][]
  align: TableAlign[]
  search: TableSearchConfig | null
}

const FENCE = /^\s*(```|~~~)/
const MARKER =
  /^\s*<nuon-table-search\b([^>]*)>(?:\s*<\/nuon-table-search>)?\s*$/
const DELIM_CELL = /^:?-+:?$/

export function parseColumns(raw?: string): string[] | null {
  if (!raw) return null
  const cols = raw
    .split(',')
    .map((c) => c.trim().toLowerCase())
    .filter(Boolean)
  return cols.length > 0 ? cols : null
}

export function makeSearchConfig(
  columns?: string,
  placeholder?: string
): TableSearchConfig {
  return {
    columns: parseColumns(columns),
    placeholder: placeholder || undefined,
  }
}

function parseMarker(attrStr: string): TableSearchConfig {
  const attrs: Record<string, string> = {}
  const re = /(\w+)="([^"]*)"/g
  let m: RegExpExecArray | null
  while ((m = re.exec(attrStr)) !== null) {
    attrs[m[1]] = m[2]
  }
  return makeSearchConfig(attrs.column ?? attrs.columns, attrs.placeholder)
}

function escapeAttr(value: string): string {
  return value.replace(/"/g, '&quot;')
}

function buildSearchAttrs(cfg: TableSearchConfig): string {
  const parts = ['data-nuon-search="1"']
  if (cfg.columns)
    parts.push(
      `data-nuon-search-columns="${escapeAttr(cfg.columns.join(','))}"`
    )
  if (cfg.placeholder)
    parts.push(`data-nuon-search-placeholder="${escapeAttr(cfg.placeholder)}"`)
  return parts.join(' ')
}

function splitRow(line: string): string[] {
  let s = line.trim()
  if (s.startsWith('|')) s = s.slice(1)
  if (s.endsWith('|')) s = s.slice(0, -1)

  const cells: string[] = []
  let current = ''
  let inCode = false

  for (let i = 0; i < s.length; i++) {
    const ch = s[i]
    if (ch === '\\' && i + 1 < s.length) {
      current += ch + s[i + 1]
      i++
      continue
    }
    if (ch === '`') {
      inCode = !inCode
      current += ch
      continue
    }
    if (ch === '|' && !inCode) {
      cells.push(current)
      current = ''
      continue
    }
    current += ch
  }
  cells.push(current)

  return cells.map((c) => c.trim())
}

function isAllDelim(cells: string[]): boolean {
  return cells.length > 0 && cells.every((c) => DELIM_CELL.test(c))
}

function parseAlign(cell: string): TableAlign {
  const c = cell.trim()
  const left = c.startsWith(':')
  const right = c.endsWith(':')
  if (left && right) return 'center'
  if (right) return 'right'
  if (left) return 'left'
  return null
}

function toText(md: string): string {
  return md
    .replace(/\\([|`*_~[\]()])/g, '$1')
    .replace(/`([^`]*)`/g, '$1')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\*([^*]+)\*/g, '$1')
    .replace(/__([^_]+)__/g, '$1')
    .replace(/_([^_]+)_/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
    .trim()
}

function toCell(md: string): TableCell {
  return { text: toText(md), markdown: md }
}

function normalizeRow(cells: string[], count: number): TableCell[] {
  const out: TableCell[] = []
  for (let k = 0; k < count; k++) {
    out.push(toCell(cells[k] ?? ''))
  }
  return out
}

export function extractTables(content: string): {
  content: string
  tableMap: Map<string, ExtractedTable>
} {
  const tableMap = new Map<string, ExtractedTable>()
  const lines = content.split('\n')
  const out: string[] = []
  let idx = 0
  let inFence = false
  let fenceMarker = ''
  let pendingSearch: TableSearchConfig | null = null

  let i = 0
  while (i < lines.length) {
    const line = lines[i]

    const fence = FENCE.exec(line)
    if (fence) {
      if (!inFence) {
        inFence = true
        fenceMarker = fence[1]
        pendingSearch = null
      } else if (fence[1] === fenceMarker) {
        inFence = false
        fenceMarker = ''
      }
      out.push(line)
      i++
      continue
    }
    if (inFence) {
      out.push(line)
      i++
      continue
    }

    const markerMatch = MARKER.exec(line)
    if (markerMatch) {
      pendingSearch = parseMarker(markerMatch[1])
      i++
      continue
    }

    if (pendingSearch && /^\s*<table[\s>]/i.test(line)) {
      out.push(
        line.replace(/<table/i, `<table ${buildSearchAttrs(pendingSearch)}`)
      )
      pendingSearch = null
      i++
      continue
    }

    const headerCells = splitRow(line)
    const delimCells = i + 1 < lines.length ? splitRow(lines[i + 1]) : []
    const isTable =
      line.includes('|') &&
      !isAllDelim(headerCells) &&
      isAllDelim(delimCells) &&
      delimCells.length === headerCells.length

    if (isTable) {
      const headers = headerCells.map(toCell)
      const align = splitRow(lines[i + 1]).map(parseAlign)
      const rows: TableCell[][] = []

      let j = i + 2
      while (
        j < lines.length &&
        lines[j].trim() !== '' &&
        lines[j].includes('|') &&
        !FENCE.test(lines[j])
      ) {
        rows.push(normalizeRow(splitRow(lines[j]), headers.length))
        j++
      }

      const id = `nuon-table-${idx++}`
      tableMap.set(id, { headers, rows, align, search: pendingSearch })
      pendingSearch = null
      out.push(`<nuon-table-rendered data-id="${id}"></nuon-table-rendered>`)
      out.push('')
      i = j
      continue
    }

    if (pendingSearch && line.trim() === '') {
      out.push(line)
      i++
      continue
    }

    pendingSearch = null
    out.push(line)
    i++
  }

  return { content: out.join('\n'), tableMap }
}
