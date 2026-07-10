import { afterEach, expect, test } from 'bun:test'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { Markdown } from './Markdown'

afterEach(cleanup)

const Surfaces = ({ children }: { children: React.ReactNode }) => (
  <MemoryRouter>
    <SurfacesProvider>{children}</SurfacesProvider>
  </MemoryRouter>
)

const TABLE = `| Name | Status | Region |
| ---- | ------ | ------ |
| prod-us | Active | us-east-1 |
| dev-eu | Inactive | eu-west-1 |`

test('table at top level renders (control)', () => {
  render(<Markdown content={TABLE} />)
  expect(screen.getByText('prod-us')).toBeInTheDocument()
  expect(screen.getByText('dev-eu')).toBeInTheDocument()
})

test('table nested inside nuon-tabs renders', () => {
  const content = `<nuon-tabs>
<nuon-tab name="clusters">
${TABLE}
</nuon-tab>
<nuon-tab name="other">
Nothing here.
</nuon-tab>
</nuon-tabs>`

  render(<Markdown content={content} />)
  expect(screen.getByText('prod-us')).toBeInTheDocument()
  expect(screen.getByText('dev-eu')).toBeInTheDocument()
})

test('table nested inside nuon-modal renders after opening', () => {
  const content = `<nuon-modal heading="Clusters" trigger="View clusters">
${TABLE}
</nuon-modal>`

  render(
    <Surfaces>
      <Markdown content={content} />
    </Surfaces>
  )

  fireEvent.click(screen.getByRole('button', { name: /view clusters/i }))
  expect(screen.getByText('prod-us')).toBeInTheDocument()
  expect(screen.getByText('dev-eu')).toBeInTheDocument()
})

test('table nested inside nuon-panel renders after opening', () => {
  const content = `<nuon-panel heading="Clusters" trigger="View clusters">
${TABLE}
</nuon-panel>`

  render(
    <Surfaces>
      <Markdown content={content} />
    </Surfaces>
  )

  fireEvent.click(screen.getByRole('button', { name: /view clusters/i }))
  expect(screen.getByText('prod-us')).toBeInTheDocument()
  expect(screen.getByText('dev-eu')).toBeInTheDocument()
})
