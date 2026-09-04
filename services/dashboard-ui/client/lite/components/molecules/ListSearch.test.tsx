import { afterEach, expect, jest, mock, test } from 'bun:test'
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { useState } from 'react'
import {
  createMemoryRouter,
  MemoryRouter,
  RouterProvider,
  useLocation,
} from 'react-router'
import { commaSetQueryParameter } from '../../lib/list-query'
import { useListQueryState } from '../../hooks/use-list-query-state'
import { ListSearch } from './ListSearch'

afterEach(() => {
  cleanup()
  jest.useRealTimers()
})

const ControlledSearch = ({ onCommit = mock() }) => {
  const [value, setValue] = useState('')

  return (
    <>
      <ListSearch
        value={value}
        onValueChange={(next) => {
          setValue(next)
          onCommit(next)
        }}
        debounceMs={300}
        aria-label="Search installs"
      />
      <output>{value}</output>
    </>
  )
}

test('commits the draft after the debounce interval', () => {
  jest.useFakeTimers()
  const onCommit = mock()
  render(<ControlledSearch onCommit={onCommit} />)

  fireEvent.change(screen.getByRole('searchbox'), {
    target: { value: 'payments' },
  })

  act(() => jest.advanceTimersByTime(299))
  expect(onCommit).not.toHaveBeenCalled()

  act(() => jest.advanceTimersByTime(1))
  expect(onCommit).toHaveBeenCalledWith('payments')
  expect(screen.getByText('payments')).toBeTruthy()
})

test('synchronizes its draft when the controlled URL value changes', () => {
  jest.useFakeTimers()
  const { rerender } = render(
    <ListSearch
      value="payments"
      onValueChange={mock()}
      aria-label="Search installs"
    />
  )

  rerender(
    <ListSearch
      value="platform"
      onValueChange={mock()}
      aria-label="Search installs"
    />
  )

  expect((screen.getByRole('searchbox') as HTMLInputElement).value).toBe(
    'platform'
  )
})

test('cancels a pending commit when unmounted', () => {
  jest.useFakeTimers()
  const onCommit = mock()
  const { unmount } = render(<ControlledSearch onCommit={onCommit} />)

  fireEvent.change(screen.getByRole('searchbox'), {
    target: { value: 'payments' },
  })
  unmount()
  act(() => jest.advanceTimersByTime(300))

  expect(onCommit).not.toHaveBeenCalled()
})

const FILTERS = {
  types: commaSetQueryParameter('types'),
}

const ListQueryExample = () => {
  const location = useLocation()
  const list = useListQueryState({ pageSize: 20, filters: FILTERS })

  return (
    <>
      <ListSearch
        value={list.search}
        onValueChange={list.setSearch}
        debounceMs={300}
        aria-label="Search installs"
      />
      <button onClick={() => list.setFilter('types', new Set(['helm']))}>
        Filter
      </button>
      <output data-testid="location">{location.search}</output>
    </>
  )
}

test('a pending search preserves a newer filter update', () => {
  jest.useFakeTimers()
  render(
    <MemoryRouter initialEntries={['/items?offset=20']}>
      <ListQueryExample />
    </MemoryRouter>
  )

  fireEvent.change(screen.getByRole('searchbox'), {
    target: { value: 'payments' },
  })
  fireEvent.click(screen.getByRole('button', { name: 'Filter' }))
  act(() => jest.advanceTimersByTime(300))

  const params = new URLSearchParams(
    screen.getByTestId('location').textContent ?? ''
  )
  expect(params.get('q')).toBe('payments')
  expect(params.get('types')).toBe('helm')
  expect(params.has('offset')).toBe(false)
})

test('Back navigation synchronizes the visible search draft', async () => {
  const router = createMemoryRouter(
    [{ path: '*', element: <ListQueryExample /> }],
    {
      initialEntries: ['/items?q=payments', '/items?q=platform'],
      initialIndex: 1,
    }
  )
  render(<RouterProvider router={router} />)

  expect((screen.getByRole('searchbox') as HTMLInputElement).value).toBe(
    'platform'
  )

  await act(() => router.navigate(-1))

  expect((screen.getByRole('searchbox') as HTMLInputElement).value).toBe(
    'payments'
  )
})
