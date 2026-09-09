import type { ColumnDef } from '@tanstack/react-table'
import { afterEach, beforeEach, describe, expect, mock, test } from 'bun:test'
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import {
  DEFAULT_USER_PREFERENCES,
  USER_PREFERENCES_STORAGE_KEY,
  UserPreferencesProvider,
} from '../../../providers/user-preferences-provider'
import { Table } from './Table'

type TRow = {
  id: string
  name: string
  status: string
}

const DATA: TRow[] = [
  { id: 'one', name: 'Production', status: 'Active' },
  { id: 'two', name: 'Staging', status: 'Deploying' },
]

const COLUMNS: ColumnDef<TRow>[] = [
  {
    accessorKey: 'name',
    header: 'Install',
    size: 300,
    cell: ({ row }) => <a href={`/${row.original.id}`}>{row.original.name}</a>,
  },
  {
    accessorKey: 'status',
    header: 'Status',
    size: 300,
  },
]

let resizeCallback: ResizeObserverCallback | undefined
const originalResizeObserver = globalThis.ResizeObserver

class MockResizeObserver {
  constructor(callback: ResizeObserverCallback) {
    resizeCallback = callback
  }

  observe() {}
  unobserve() {}
  disconnect() {}
}

const setWidth = (width: number) => {
  act(() => {
    resizeCallback?.(
      [{ contentRect: { width } } as ResizeObserverEntry],
      {} as ResizeObserver
    )
  })
}

const Example = ({
  data = DATA,
  loading,
  toolbar,
  renderCard = ({ row }: { row: { original: TRow } }) => (
    <article>{row.original.name} card</article>
  ),
}: {
  data?: TRow[]
  loading?: boolean
  toolbar?: React.ReactNode
  renderCard?: ({ row }: { row: { original: TRow } }) => React.ReactNode
}) => (
  <UserPreferencesProvider>
    <Table
      data={data}
      columns={COLUMNS}
      getRowId={(row) => row.id}
      loading={loading}
      toolbar={toolbar}
      emptyState="No installs yet"
      renderCard={renderCard}
    />
  </UserPreferencesProvider>
)

beforeEach(() => {
  window.localStorage.clear()
  window.sessionStorage.clear()
  resizeCallback = undefined
  Object.defineProperty(globalThis, 'ResizeObserver', {
    value: MockResizeObserver,
    writable: true,
    configurable: true,
  })
})

afterEach(() => {
  cleanup()
  window.localStorage.clear()
  window.sessionStorage.clear()
  Object.defineProperty(globalThis, 'ResizeObserver', {
    value: originalResizeObserver,
    writable: true,
    configurable: true,
  })
})

describe('Table', () => {
  test('renders semantic headers, rows, and cells', () => {
    render(<Example />)
    setWidth(800)

    expect(screen.getByRole('table')).toBeTruthy()
    expect(
      screen.getAllByRole('columnheader').map((header) => header.textContent)
    ).toEqual(['Install', 'Status'])
    expect(screen.getAllByRole('row')).toHaveLength(3)
    expect(screen.getByRole('link', { name: 'Production' })).toBeTruthy()
  })

  test('passes the same TanStack row to the required card renderer', () => {
    const renderCard = mock(({ row }) => (
      <article>{row.original.name} card</article>
    ))
    render(<Example renderCard={renderCard} />)
    setWidth(800)

    fireEvent.click(screen.getByRole('button', { name: 'Card view' }))

    expect(screen.queryByRole('table')).toBeNull()
    expect(screen.getByText('Production card')).toBeTruthy()
    expect(renderCard.mock.calls[0][0].row.original === DATA[0]).toBe(true)
  })

  test('aligns toolbar controls with the view toggle', () => {
    const { container } = render(
      <Example toolbar={<input aria-label="Search installs" />} />
    )
    setWidth(800)

    const toolbarRow = container.querySelector('[data-table-toolbar]')
    expect(toolbarRow).not.toBeNull()
    expect(toolbarRow?.contains(screen.getByLabelText('Search installs'))).toBe(
      true
    )
    expect(
      toolbarRow?.contains(
        screen.getByRole('group', { name: 'Collection view' })
      )
    ).toBe(true)

    setWidth(500)
    expect(screen.getByLabelText('Search installs')).toBeTruthy()
    expect(screen.queryByRole('group', { name: 'Collection view' })).toBeNull()
  })

  test('forces cards when columns do not fit and restores the preference', () => {
    render(<Example />)

    setWidth(500)
    expect(screen.queryByRole('table')).toBeNull()
    expect(screen.queryByRole('button', { name: 'Table view' })).toBeNull()

    setWidth(800)
    expect(screen.getByRole('table')).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Table view' })).toBeTruthy()
  })

  test('does not overwrite a card preference while cards are forced', () => {
    render(<Example />)
    setWidth(800)
    fireEvent.click(screen.getByRole('button', { name: 'Card view' }))

    setWidth(500)
    setWidth(800)

    expect(screen.queryByRole('table')).toBeNull()
    expect(
      screen
        .getByRole('button', { name: 'Card view' })
        .getAttribute('aria-pressed')
    ).toBe('true')
  })

  test('keeps page view changes local and reapplies the saved default', () => {
    window.localStorage.setItem(
      USER_PREFERENCES_STORAGE_KEY,
      JSON.stringify({
        version: 1,
        preferences: {
          ...DEFAULT_USER_PREFERENCES,
          collectionView: 'cards',
        },
      })
    )

    const first = render(<Example />)
    setWidth(800)
    expect(screen.queryByRole('table')).toBeNull()

    fireEvent.click(screen.getByRole('button', { name: 'Table view' }))
    expect(screen.getByRole('table')).toBeTruthy()
    expect(
      JSON.parse(
        window.localStorage.getItem(USER_PREFERENCES_STORAGE_KEY) ?? ''
      ).preferences.collectionView
    ).toBe('cards')

    first.unmount()

    render(<Example />)
    setWidth(800)

    expect(screen.queryByRole('table')).toBeNull()
    expect(screen.getByText('Production card')).toBeTruthy()
  })

  test('keeps view switching usable when local storage is blocked', () => {
    const descriptor = Object.getOwnPropertyDescriptor(window, 'localStorage')
    Object.defineProperty(window, 'localStorage', {
      configurable: true,
      get: () => {
        throw new Error('blocked')
      },
    })

    try {
      render(<Example />)
      setWidth(800)
      fireEvent.click(screen.getByRole('button', { name: 'Card view' }))
      expect(screen.getByText('Production card')).toBeTruthy()
    } finally {
      if (descriptor) {
        Object.defineProperty(window, 'localStorage', descriptor)
      }
    }
  })

  test('uses stable resource row identities in both presentations', () => {
    const { container } = render(<Example />)
    setWidth(800)
    expect(container.querySelector('tr[data-row-id="one"]')).not.toBeNull()

    fireEvent.click(screen.getByRole('button', { name: 'Card view' }))
    expect(
      container.querySelector('[data-table-card][data-row-id="one"]')
    ).not.toBeNull()
  })

  test('renders loading and empty states inside the active presentation', () => {
    const { container, rerender } = render(<Example loading />)
    setWidth(800)
    expect(screen.getByRole('status', { name: 'Loading results' })).toBeTruthy()

    rerender(<Example data={[]} />)
    expect(screen.getByText('No installs yet')).toBeTruthy()
    expect(container.querySelector('[data-table-loading-row]')).toBeNull()
  })

  test('loads as skeleton rows that keep the real table chrome', () => {
    const { container } = render(<Example loading />)
    setWidth(800)

    const loadingRows = container.querySelectorAll('[data-table-loading-row]')

    expect(screen.getByRole('columnheader', { name: 'Install' })).toBeTruthy()
    expect(loadingRows.length).toBe(5)
    expect(container.querySelectorAll('.skeleton-text').length).toBe(15)
  })

  test('reserves the identity cell name and ID lines while loading', () => {
    const { container } = render(<Example loading />)
    setWidth(800)

    const [identityCell, statusCell] = [
      ...(container
        .querySelector('[data-table-loading-row]')
        ?.querySelectorAll('td') ?? []),
    ]

    expect(identityCell?.querySelectorAll('.skeleton-text').length).toBe(2)
    expect(statusCell?.querySelectorAll('.skeleton-text').length).toBe(1)
  })

  test('loads as skeleton cards in card view', () => {
    const { container } = render(<Example loading />)
    setWidth(800)
    fireEvent.click(screen.getByRole('button', { name: 'Card view' }))

    expect(container.querySelectorAll('[data-table-loading-card]').length).toBe(
      5
    )
    expect(screen.getAllByText('Status').length).toBe(5)
  })
})
