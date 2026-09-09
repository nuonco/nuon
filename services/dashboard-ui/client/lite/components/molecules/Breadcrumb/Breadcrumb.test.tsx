import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { Breadcrumb, hiddenBreadcrumbIndices } from './Breadcrumb'

afterEach(cleanup)

const renderBreadcrumb = (
  items: Parameters<typeof Breadcrumb>[0]['items']
) =>
  render(
    <MemoryRouter>
      <Breadcrumb items={items} />
    </MemoryRouter>
  )

describe('Breadcrumb', () => {
  test('renders ancestors as links and the current page as text', () => {
    renderBreadcrumb([
      { label: 'Acme', href: '/org' },
      { label: 'Installs', href: '/org/installs' },
      { label: 'Production' },
    ])

    expect(screen.getByRole('navigation', { name: 'Breadcrumb' })).toBeTruthy()
    expect(screen.getByRole('link', { name: 'Acme' }).getAttribute('href')).toBe(
      '/org'
    )
    expect(screen.getByRole('link', { name: 'Installs' })).toBeTruthy()
    expect(screen.queryByRole('link', { name: 'Production' })).toBeNull()
    const visibleTrail = screen
      .getByRole('navigation', { name: 'Breadcrumb' })
      .querySelector<HTMLOListElement>('ol:not([aria-hidden])')!
    expect(
      within(visibleTrail).getByText('Production').getAttribute('aria-current')
    ).toBe('page')
  })

  test('keeps unresolved segments in the loading state', () => {
    const { container } = renderBreadcrumb([
      { href: '/org', loadingWidth: 16 },
      { label: 'Installs', href: '/org/installs' },
      { loadingWidth: 12 },
    ])

    const visibleTrail = container.querySelector('ol:not([aria-hidden])')!
    expect(visibleTrail.querySelectorAll('.skeleton-text')).toHaveLength(2)
    expect(visibleTrail.querySelectorAll('li')).toHaveLength(3)
  })

  test('hides middle ancestors before the first ancestor', () => {
    expect(hiddenBreadcrumbIndices(300, [70, 80, 90, 70])).toEqual([1])
    expect(hiddenBreadcrumbIndices(200, [70, 80, 90, 70])).toEqual([1, 2])
    expect(hiddenBreadcrumbIndices(150, [70, 80, 90, 70])).toEqual([
      0, 1, 2,
    ])
  })

  test('keeps every segment when they fit', () => {
    expect(hiddenBreadcrumbIndices(400, [70, 80, 90, 70])).toEqual([])
  })
})
