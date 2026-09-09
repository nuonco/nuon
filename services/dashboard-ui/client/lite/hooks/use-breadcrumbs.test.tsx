import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { BreadcrumbProvider } from '../providers/breadcrumb-provider'
import { useBreadcrumbItems, useBreadcrumbs } from './use-breadcrumbs'

afterEach(cleanup)

const Trail = () => {
  const items = useBreadcrumbItems()
  return (
    <span data-testid="trail">
      {items.map((item) => item.label ?? 'loading').join(' / ')}
    </span>
  )
}

const Child = ({ label }: { label?: string }) => {
  useBreadcrumbs([
    { label: 'Acme', href: '/org' },
    { label: 'Apps', href: '/org/apps' },
    { label },
  ])
  return null
}

test('clears the active trail when its route unmounts', async () => {
  const view = render(
    <BreadcrumbProvider>
      <Trail />
      <Child label="Payments" />
    </BreadcrumbProvider>
  )

  await waitFor(() =>
    expect(screen.getByText('Acme / Apps / Payments')).toBeTruthy()
  )

  view.rerender(
    <BreadcrumbProvider>
      <Trail />
    </BreadcrumbProvider>
  )

  await waitFor(() => expect(screen.getByTestId('trail').textContent).toBe(''))
})

test('updates unresolved labels without changing the trail shape', async () => {
  const view = render(
    <BreadcrumbProvider>
      <Trail />
      <Child />
    </BreadcrumbProvider>
  )

  await waitFor(() =>
    expect(screen.getByText('Acme / Apps / loading')).toBeTruthy()
  )

  view.rerender(
    <BreadcrumbProvider>
      <Trail />
      <Child label="Payments" />
    </BreadcrumbProvider>
  )

  await waitFor(() =>
    expect(screen.getByText('Acme / Apps / Payments')).toBeTruthy()
  )
})
