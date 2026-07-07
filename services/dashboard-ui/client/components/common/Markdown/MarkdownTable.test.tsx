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
