import type { ColumnDef } from '@tanstack/react-table'
import { afterEach, beforeEach, describe, expect, mock, test } from 'bun:test'
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { TABLE_VIEW_STORAGE_KEY } from '../../../hooks/use-table-view'
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
  <Table
    data={data}
    columns={COLUMNS}
    getRowId={(row) => row.id}
    loading={loading}
    toolbar={toolbar}
    emptyState="No installs yet"
    renderCard={renderCard}
  />
)

beforeEach(() => {
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
    expect(toolbarRow?.contains(screen.getByRole('group', { name: 'Collection view' }))).toBe(
      true
    )

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

  test('shares the preferred view through session storage', () => {
    const first = render(<Example />)
    setWidth(800)
    fireEvent.click(screen.getByRole('button', { name: 'Card view' }))
    expect(window.sessionStorage.getItem(TABLE_VIEW_STORAGE_KEY)).toBe('cards')
    first.unmount()

    render(<Example />)
    setWidth(800)

    expect(screen.queryByRole('table')).toBeNull()
    expect(screen.getByText('Production card')).toBeTruthy()
  })

  test('keeps view switching usable when session storage is blocked', () => {
    window.sessionStorage.setItem(TABLE_VIEW_STORAGE_KEY, 'table')
    const descriptor = Object.getOwnPropertyDescriptor(window, 'sessionStorage')
    Object.defineProperty(window, 'sessionStorage', {
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
        Object.defineProperty(window, 'sessionStorage', descriptor)
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

    expect(screen.getByRole('columnheader', { name: 'Install' })).toBeTruthy()
    expect(container.querySelectorAll('[data-table-loading-row]').length).toBe(
      5
    )
    expect(container.querySelectorAll('.skeleton-text').length).toBe(10)
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
