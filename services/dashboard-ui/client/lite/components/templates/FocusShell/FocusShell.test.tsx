import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter, Outlet, Route, Routes } from 'react-router'
import { FocusShell } from './FocusShell'

afterEach(cleanup)

test('renders sticky global chrome inside the single scroll region', () => {
  const { container } = render(
    <MemoryRouter>
      <FocusShell actions={<button type="button">Account</button>}>
        <span>Focused content</span>
      </FocusShell>
    </MemoryRouter>
  )

  const header = screen.getByRole('banner')
  const main = screen.getByRole('main')

  expect(screen.getByRole('link', { name: 'Nuon home' })).toBeTruthy()
  expect(screen.getByRole('button', { name: 'Account' })).toBeTruthy()
  expect(header.parentElement).toBe(main)
  expect(main.firstElementChild).toBe(header)
  expect(container.querySelectorAll('.overflow-y-auto')).toHaveLength(1)
  expect(screen.queryByRole('complementary')).toBeNull()
  expect(screen.queryByRole('contentinfo')).toBeNull()
  expect(
    container
      .querySelector('[data-shell-background]')
      ?.getAttribute('aria-hidden')
  ).toBe('true')
})

test('centers and gutters content by default', () => {
  render(
    <MemoryRouter>
      <FocusShell>
        <span>Focused content</span>
      </FocusShell>
    </MemoryRouter>
  )

  const frame = screen.getByRole('main').children.item(1)

  expect(frame?.classList.contains('max-w-6xl')).toBe(true)
  expect(frame?.classList.contains('mx-auto')).toBe(true)
  expect(frame?.classList.contains('px-4')).toBe(true)
})

test('supports full-bleed focused content', () => {
  render(
    <MemoryRouter>
      <FocusShell fullBleed>
        <span>Focused content</span>
      </FocusShell>
    </MemoryRouter>
  )

  const frame = screen.getByRole('main').children.item(1)

  expect(frame?.classList.contains('w-full')).toBe(true)
  expect(frame?.classList.contains('max-w-6xl')).toBe(false)
  expect(frame?.classList.contains('px-4')).toBe(false)
})

test('adds glass elevation after focused content scrolls', () => {
  render(
    <MemoryRouter>
      <FocusShell>
        <span>Focused content</span>
      </FocusShell>
    </MemoryRouter>
  )

  const main = screen.getByRole('main')
  const header = screen.getByRole('banner')

  expect(header.classList.contains('!bg-transparent')).toBe(true)
  expect(header.classList.contains('shadow-none')).toBe(true)

  Object.defineProperty(main, 'scrollTop', {
    configurable: true,
    value: 12,
  })
  fireEvent.scroll(main)

  expect(header.classList.contains('!bg-transparent')).toBe(false)
  expect(
    header.classList.contains('shadow-[var(--card-shadow-floating)]')
  ).toBe(true)
})

test('composes as a nested route layout', () => {
  render(
    <MemoryRouter initialEntries={['/setup']}>
      <Routes>
        <Route
          element={
            <FocusShell>
              <Outlet />
            </FocusShell>
          }
        >
          <Route path="/setup" element={<h1>Account setup</h1>} />
        </Route>
      </Routes>
    </MemoryRouter>
  )

  expect(screen.getByRole('heading', { name: 'Account setup' })).toBeTruthy()
  expect(screen.getByRole('main')).toBeTruthy()
})
