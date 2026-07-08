import { expect, test } from 'bun:test'
import { extractTables, type ExtractedTable } from './markdown-table'

function onlyTable(content: string): ExtractedTable {
  const { tableMap } = extractTables(content)
  const tables = [...tableMap.values()]
  expect(tables.length).toBe(1)
  return tables[0]
}

test('parses a basic pipe table into a placeholder + structured data', () => {
  const { content, tableMap } = extractTables(
    `| Name | Status |
| ---- | ------ |
| prod | live   |
| dev  | error  |`
  )

  expect(content).toContain('<nuon-table-rendered data-id="nuon-table-0">')
  const table = tableMap.get('nuon-table-0')!
  expect(table.headers.map((h) => h.text)).toEqual(['Name', 'Status'])
  expect(table.rows.length).toBe(2)
  expect(table.rows[0].map((c) => c.text)).toEqual(['prod', 'live'])
  expect(table.rows[1].map((c) => c.text)).toEqual(['dev', 'error'])
  expect(table.search).toBeNull()
})

test('associates a preceding nuon-table-search marker with the table', () => {
  const table = onlyTable(
    `<nuon-table-search column="name"></nuon-table-search>

| Name | Status |
| ---- | ------ |
| prod | live   |`
  )
  expect(table.search).toEqual({ columns: ['name'], placeholder: undefined })
})

test('search marker supports comma-separated columns', () => {
  const table = onlyTable(
    `<nuon-table-search column="name, status">

| Name | Status |
| ---- | ------ |
| prod | live   |`
  )
  expect(table.search?.columns).toEqual(['name', 'status'])
})

test('search marker with no column searches all columns (null)', () => {
  const table = onlyTable(
    `<nuon-table-search>

| Name | Status |
| ---- | ------ |
| prod | live   |`
  )
  expect(table.search).not.toBeNull()
  expect(table.search?.columns).toBeNull()
})

test('marker not immediately followed by a table is dropped', () => {
  const { tableMap } = extractTables(
    `<nuon-table-search column="name"></nuon-table-search>

Some paragraph in between.

| Name | Status |
| ---- | ------ |
| prod | live   |`
  )
  const table = [...tableMap.values()][0]
  expect(table.search).toBeNull()
})

test('respects escaped pipes inside cells', () => {
  const table = onlyTable(
    `| Command | Note |
| ------- | ---- |
| a \\| b  | pipe |`
  )
  expect(table.rows[0].map((c) => c.text)).toEqual(['a | b', 'pipe'])
})

test('respects pipes inside inline code spans', () => {
  const table = onlyTable(
    `| Command | Note |
| ------- | ---- |
| \`a | b\` | code |`
  )
  expect(table.rows[0].length).toBe(2)
  expect(table.rows[0][1].text).toBe('code')
})

test('does not parse pipe tables inside fenced code blocks', () => {
  const { tableMap } = extractTables(
    `\`\`\`
| Name | Status |
| ---- | ------ |
| prod | live   |
\`\`\``
  )
  expect(tableMap.size).toBe(0)
})

test('does not treat a setext heading with a pipe as a table', () => {
  const { tableMap } = extractTables(
    `Some | heading
---

body text`
  )
  expect(tableMap.size).toBe(0)
})

test('parses column alignment from the delimiter row', () => {
  const table = onlyTable(
    `| L | C | R |
| :-- | :-: | --: |
| a | b | c |`
  )
  expect(table.align).toEqual(['left', 'center', 'right'])
})

test('strips inline markdown for the searchable text but keeps raw markdown', () => {
  const table = onlyTable(
    `| Name | Link |
| ---- | ---- |
| **bold** | [docs](https://x.dev) |`
  )
  expect(table.rows[0][0].text).toBe('bold')
  expect(table.rows[0][0].markdown).toBe('**bold**')
  expect(table.rows[0][1].text).toBe('docs')
  expect(table.rows[0][1].markdown).toBe('[docs](https://x.dev)')
})

test('normalizes ragged rows to the header column count', () => {
  const table = onlyTable(
    `| A | B | C |
| - | - | - |
| 1 | 2 |
| 1 | 2 | 3 | 4 |`
  )
  expect(table.rows[0].map((c) => c.text)).toEqual(['1', '2', ''])
  expect(table.rows[1].map((c) => c.text)).toEqual(['1', '2', '3'])
})

test('injects search attributes into a marked HTML table', () => {
  const { content, tableMap } = extractTables(
    `<nuon-table-search column="monitor"></nuon-table-search>

<table>
<thead><tr><th>Monitor</th></tr></thead>
<tbody><tr><td>api</td></tr></tbody>
</table>`
  )
  expect(tableMap.size).toBe(0)
  expect(content).toContain('data-nuon-search="1"')
  expect(content).toContain('data-nuon-search-columns="monitor"')
  expect(content).not.toContain('<nuon-table-search')
})

test('injects placeholder and preserves existing table attributes', () => {
  const { content } = extractTables(
    `<nuon-table-search placeholder="Find a monitor…"></nuon-table-search>

<table style="width:100%">
<tr><td>api</td></tr>
</table>`
  )
  expect(content).toContain('data-nuon-search-placeholder="Find a monitor…"')
  expect(content).toContain('style="width:100%"')
  expect(content).not.toContain('data-nuon-search-columns')
})

test('leaves unmarked HTML tables untouched', () => {
  const html = `<table>
<tr><td>api</td></tr>
</table>`
  const { content, tableMap } = extractTables(html)
  expect(tableMap.size).toBe(0)
  expect(content).toBe(html)
  expect(content).not.toContain('data-nuon-search')
})

test('extracts multiple tables independently', () => {
  const { tableMap } = extractTables(
    `| A |
| - |
| 1 |

<nuon-table-search column="b"></nuon-table-search>

| B |
| - |
| 2 |`
  )
  expect(tableMap.size).toBe(2)
  expect(tableMap.get('nuon-table-0')?.search).toBeNull()
  expect(tableMap.get('nuon-table-1')?.search?.columns).toEqual(['b'])
})
