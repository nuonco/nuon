import { afterEach, expect, test } from 'bun:test'
import { render, screen, fireEvent, within, cleanup } from '@testing-library/react'
import { Markdown } from './Markdown'

afterEach(cleanup)

const CLUSTER_TABLE = `| Name | Status | Region |
| ---- | ------ | ------ |
| prod-us | Active | us-east-1 |
| dev-eu | Inactive | eu-west-1 |
| staging-us | Active | us-west-2 |`

function rowText(): string[] {
  return screen
    .getAllByRole('row')
    .filter((r) => within(r).queryAllByRole('cell').length > 1)
    .map((r) => r.textContent ?? '')
}

test('renders a markdown pipe table through the common Table', () => {
  render(<Markdown content={CLUSTER_TABLE} />)
  expect(screen.getByText('prod-us')).toBeInTheDocument()
  expect(screen.getByText('dev-eu')).toBeInTheDocument()
  expect(rowText().length).toBe(3)
})

test('renders no search box without a marker', () => {
  render(<Markdown content={CLUSTER_TABLE} />)
  expect(screen.queryByPlaceholderText(/search/i)).toBeNull()
})

test('filters rows client-side by the configured column', () => {
  render(
    <Markdown
      content={`<nuon-table-search column="name"></nuon-table-search>

${CLUSTER_TABLE}`}
    />
  )

  const input = screen.getByPlaceholderText(/search/i)
  fireEvent.change(input, { target: { value: 'prod' } })

  const rows = rowText()
  expect(rows.length).toBe(1)
  expect(rows[0]).toContain('prod-us')
})

test('column filter ignores values in other columns', () => {
  render(
    <Markdown
      content={`<nuon-table-search column="name"></nuon-table-search>

${CLUSTER_TABLE}`}
    />
  )

  const input = screen.getByPlaceholderText(/search/i)
  fireEvent.change(input, { target: { value: 'active' } })

  expect(rowText().length).toBe(0)
})

test('all-column search matches a non-name column', () => {
  render(
    <Markdown
      content={`<nuon-table-search></nuon-table-search>

${CLUSTER_TABLE}`}
    />
  )

  const input = screen.getByPlaceholderText(/search/i)
  fireEvent.change(input, { target: { value: 'us-west-2' } })

  const rows = rowText()
  expect(rows.length).toBe(1)
  expect(rows[0]).toContain('staging-us')
})

const HTML_TABLE = `<table>
<thead>
<tr><th>Monitor</th><th>Status</th></tr>
</thead>
<tbody>
<tr><td>api-healthcheck</td><td>finished</td></tr>
<tr><td>db-healthcheck</td><td>error</td></tr>
<tr><td>cache-healthcheck</td><td>running</td></tr>
</tbody>
</table>`

test('renders an unmarked HTML table through the common Table (sortable, no search box)', () => {
  render(<Markdown content={HTML_TABLE} />)
  expect(screen.getByText('api-healthcheck')).toBeInTheDocument()
  expect(screen.queryByPlaceholderText(/search/i)).toBeNull()

  const header = screen.getByText('Monitor').closest('th')
  expect(header).not.toBeNull()
  fireEvent.click(header!)
  const rows = rowText()
  expect(rows.length).toBe(3)
  expect(rows[0]).toContain('api-healthcheck')
})

test('converts a marked HTML table into a searchable table', () => {
  render(
    <Markdown
      content={`<nuon-table-search column="monitor"></nuon-table-search>

${HTML_TABLE}`}
    />
  )

  expect(screen.getByPlaceholderText(/search/i)).toBeInTheDocument()
  expect(rowText().length).toBe(3)

  const input = screen.getByPlaceholderText(/search/i)
  fireEvent.change(input, { target: { value: 'db' } })

  const rows = rowText()
  expect(rows.length).toBe(1)
  expect(rows[0]).toContain('db-healthcheck')
})

test('falls back to plain rendering for HTML tables with colspan', () => {
  render(
    <Markdown
      content={`<nuon-table-search column="monitor"></nuon-table-search>

<table>
<thead>
<tr><th>Monitor</th><th>Status</th></tr>
</thead>
<tbody>
<tr><td colspan="2">grouped row</td></tr>
<tr><td>api</td><td>ok</td></tr>
</tbody>
</table>`}
    />
  )

  expect(screen.getByText('grouped row')).toBeInTheDocument()
  expect(screen.queryByPlaceholderText(/search/i)).toBeNull()
})

test('sorts a time column by timestamp, not by rendered relative text', () => {
  render(
    <Markdown
      content={`<table>
<thead>
<tr><th>Name</th><th>Updated</th></tr>
</thead>
<tbody>
<tr><td>bravo</td><td><nuon-time time="2026-07-20T00:00:00Z" format="relative"></nuon-time></td></tr>
<tr><td>alpha</td><td><nuon-time time="2026-07-01T00:00:00Z" format="relative"></nuon-time></td></tr>
<tr><td>charlie</td><td><nuon-time time="2026-07-10T00:00:00Z" format="relative"></nuon-time></td></tr>
</tbody>
</table>`}
    />
  )

  const header = screen.getByText('Updated').closest('th')
  expect(header).not.toBeNull()
  fireEvent.click(header!)

  const rows = rowText()
  expect(rows.map((r) => r.match(/alpha|bravo|charlie/)?.[0])).toEqual([
    'alpha',
    'charlie',
    'bravo',
  ])
})

test('renders formatted cell content and matches on its text', () => {
  render(
    <Markdown
      content={`<nuon-table-search column="policy"></nuon-table-search>

| Policy | Docs |
| ------ | ---- |
| **Billing** change | [runbook](https://docs.nuon.co/billing) |
| Namespace limit | [runbook](https://docs.nuon.co/ns) |`}
    />
  )

  const link = screen.getAllByRole('link')[0]
  expect(link).toHaveAttribute('href', 'https://docs.nuon.co/billing')

  const input = screen.getByPlaceholderText(/search/i)
  fireEvent.change(input, { target: { value: 'billing' } })

  const rows = rowText()
  expect(rows.length).toBe(1)
  expect(rows[0]).toContain('Billing')
})
