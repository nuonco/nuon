import { afterEach, beforeEach, expect, mock, test } from 'bun:test'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { MemoryRouter, useLocation } from 'react-router'
import {
  DASHBOARD_SIDEBAR_STORAGE_KEY,
  DashboardShellProvider,
  useDashboardShell,
} from '../../../providers/dashboard-shell-provider'
import { useNavShortcuts } from '../../../hooks/use-nav-shortcuts'
import type { INavItem } from '../../molecules/NavLink'
import { DashboardShell } from './DashboardShell'

const originalMatchMedia = window.matchMedia
let desktop = true
let mediaListeners = new Set<(event: MediaQueryListEvent) => void>()

const installMatchMedia = (matches: boolean) => {
  desktop = matches
  mediaListeners = new Set()
  window.matchMedia = mock(
    () =>
      ({
        matches: desktop,
        media: '(min-width: 768px)',
        onchange: null,
        addEventListener: (
          _type: string,
          listener: (event: MediaQueryListEvent) => void
        ) => mediaListeners.add(listener),
        removeEventListener: (
          _type: string,
          listener: (event: MediaQueryListEvent) => void
        ) => mediaListeners.delete(listener),
        addListener: () => {},
        removeListener: () => {},
        dispatchEvent: () => true,
      }) as MediaQueryList
  )
}

beforeEach(() => {
  localStorage.clear()
  installMatchMedia(true)
})

afterEach(() => {
  cleanup()
  window.matchMedia = originalMatchMedia
})

const ShellState = () => {
  const { desktopSidebarExpanded, mobileSidebarOpen, toggleSidebar } =
    useDashboardShell()

  return (
    <>
      <span data-testid="desktop-state">
        {desktopSidebarExpanded ? 'expanded' : 'collapsed'}
      </span>
      <span data-testid="mobile-state">
        {mobileSidebarOpen ? 'open' : 'closed'}
      </span>
      <button type="button" onClick={toggleSidebar}>
        Toggle
      </button>
    </>
  )
}

test('persists desktop expansion before the next mount', () => {
  const first = render(
    <DashboardShellProvider>
      <ShellState />
    </DashboardShellProvider>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Toggle' }))
  expect(screen.getByTestId('desktop-state').textContent).toBe('collapsed')
  expect(localStorage.getItem(DASHBOARD_SIDEBAR_STORAGE_KEY)).toBe('false')

  first.unmount()
  render(
    <DashboardShellProvider>
      <ShellState />
    </DashboardShellProvider>
  )

  expect(screen.getByTestId('desktop-state').textContent).toBe('collapsed')
})

test('keeps mobile drawer state transient', () => {
  installMatchMedia(false)
  render(
    <DashboardShellProvider>
      <ShellState />
    </DashboardShellProvider>
  )

  fireEvent.click(screen.getByRole('button', { name: 'Toggle' }))

  expect(screen.getByTestId('mobile-state').textContent).toBe('open')
  expect(localStorage.getItem(DASHBOARD_SIDEBAR_STORAGE_KEY)).toBeNull()
})

test('toggles the active sidebar mode with Alt+S', () => {
  render(
    <DashboardShellProvider>
      <ShellState />
    </DashboardShellProvider>
  )

  fireEvent.keyDown(window, { key: 's', code: 'KeyS', altKey: true })

  expect(screen.getByTestId('desktop-state').textContent).toBe('collapsed')
})

const NAVIGATION: INavItem[] = [
  {
    href: '/',
    label: 'Dashboard',
    icon: 'HouseIcon',
    shortcut: 'g d',
    end: true,
  },
  {
    href: '/apps',
    label: 'Apps',
    icon: 'AppWindowIcon',
    shortcut: 'g a',
  },
]

test('keeps navigation labels mounted through sidebar transitions', () => {
  render(
    <MemoryRouter>
      <DashboardShell primaryNav={NAVIGATION} initialDesktopExpanded={false}>
        <span>Content</span>
      </DashboardShell>
    </MemoryRouter>
  )

  const link = screen.getByRole('link', { name: 'Apps' })
  const content = link.children.item(1)

  expect(content?.getAttribute('aria-hidden')).toBe('true')
  fireEvent.click(screen.getByRole('button', { name: 'Expand sidebar' }))
  expect(content?.isConnected).toBe(true)
  expect(content?.hasAttribute('aria-hidden')).toBe(false)
})

const CurrentPath = () => {
  const location = useLocation()
  return <span data-testid="path">{location.pathname}</span>
}

const ShortcutHarness = () => {
  useNavShortcuts(NAVIGATION)
  return (
    <>
      <input aria-label="Editor" />
      <CurrentPath />
    </>
  )
}

test('navigates from the displayed chord and ignores editable targets', () => {
  render(
    <MemoryRouter initialEntries={['/']}>
      <ShortcutHarness />
    </MemoryRouter>
  )

  fireEvent.keyDown(document, { key: 'g' })
  fireEvent.keyDown(document, { key: 'a' })
  expect(screen.getByTestId('path').textContent).toBe('/apps')

  const input = screen.getByRole('textbox', { name: 'Editor' })
  fireEvent.keyDown(input, { key: 'g' })
  fireEvent.keyDown(input, { key: 'd' })
  expect(screen.getByTestId('path').textContent).toBe('/apps')
})

test('closes mobile navigation after selecting an internal route', async () => {
  installMatchMedia(false)
  render(
    <MemoryRouter initialEntries={['/']}>
      <DashboardShell
        primaryNav={NAVIGATION}
        userMenu={<span>Account control</span>}
      >
        <CurrentPath />
      </DashboardShell>
    </MemoryRouter>
  )

  const trigger = screen.getByRole('button', { name: 'Open navigation' })
  fireEvent.click(trigger)
  expect(screen.getByRole('dialog', { name: 'Main navigation' })).toBeTruthy()
  expect(screen.getAllByText('Account control')).toHaveLength(1)

  fireEvent.click(screen.getByRole('link', { name: /Apps/ }))

  expect(screen.queryByRole('dialog', { name: 'Main navigation' })).toBeNull()
  expect(screen.getByTestId('path').textContent).toBe('/apps')
  await waitFor(() => expect(document.activeElement).toBe(trigger))
})

test('closes mobile navigation with Escape and restores focus', async () => {
  installMatchMedia(false)
  render(
    <MemoryRouter>
      <DashboardShell primaryNav={NAVIGATION}>
        <span>Content</span>
      </DashboardShell>
    </MemoryRouter>
  )

  const trigger = screen.getByRole('button', { name: 'Open navigation' })
  fireEvent.click(trigger)
  const dialog = screen.getByRole('dialog', { name: 'Main navigation' })

  fireEvent.keyDown(dialog, { key: 'Escape' })

  expect(screen.queryByRole('dialog', { name: 'Main navigation' })).toBeNull()
  await waitFor(() => expect(document.activeElement).toBe(trigger))
})

test('contains mobile focus and closes from the backdrop', async () => {
  installMatchMedia(false)
  const { container } = render(
    <MemoryRouter>
      <DashboardShell primaryNav={NAVIGATION}>
        <span>Content</span>
      </DashboardShell>
    </MemoryRouter>
  )

  const trigger = screen.getByRole('button', { name: 'Open navigation' })
  fireEvent.click(trigger)
  const dialog = screen.getByRole('dialog', { name: 'Main navigation' })
  const focusable = Array.from(
    dialog.querySelectorAll<HTMLElement>('a[href], button')
  )
  const first = focusable[0]
  const last = focusable.at(-1)

  last?.focus()
  fireEvent.keyDown(dialog, { key: 'Tab' })
  expect(document.activeElement).toBe(first)

  first?.focus()
  fireEvent.keyDown(dialog, { key: 'Tab', shiftKey: true })
  expect(document.activeElement).toBe(last)

  const backdrop = container.querySelector<HTMLElement>(
    '[data-dashboard-backdrop]'
  )
  fireEvent.click(backdrop!)

  expect(screen.queryByRole('dialog', { name: 'Main navigation' })).toBeNull()
  await waitFor(() => expect(document.activeElement).toBe(trigger))
})

test('adds glass chrome after content scrolls', () => {
  const { container } = render(
    <MemoryRouter>
      <DashboardShell
        primaryNav={NAVIGATION}
        statusBar={<span>Connected</span>}
      >
        <span>Content</span>
      </DashboardShell>
    </MemoryRouter>
  )

  const scrollRegion = container.querySelector<HTMLElement>(
    '[data-dashboard-scroll]'
  )
  const header = screen.getByRole('banner')
  const sidebar = screen.getByRole('complementary')
  const statusBar = screen.getByRole('contentinfo')

  expect(container.querySelector('[data-shell-background]')).toBeTruthy()
  expect(header.classList.contains('!bg-transparent')).toBe(true)
  expect(header.classList.contains('shadow-none')).toBe(true)
  expect(
    sidebar.classList.contains('shadow-[var(--card-shadow-floating)]')
  ).toBe(true)
  expect(
    statusBar.classList.contains('shadow-[var(--card-shadow-floating)]')
  ).toBe(true)

  Object.defineProperty(scrollRegion!, 'scrollTop', {
    configurable: true,
    value: 12,
  })
  fireEvent.scroll(scrollRegion!)

  expect(header.classList.contains('!bg-transparent')).toBe(false)
  expect(
    header.classList.contains('shadow-[var(--card-shadow-floating)]')
  ).toBe(true)
  expect(statusBar.textContent).toContain('Connected')
})
