import { afterEach, expect, test } from 'bun:test'
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider, useLocation } from 'react-router'
import {
  booleanQueryParameter,
  commaSetQueryParameter,
} from '../lib/list-query'
import { useListQueryState } from './use-list-query-state'

afterEach(cleanup)

const FILTERS = {
  active: booleanQueryParameter('active'),
  types: commaSetQueryParameter<'helm' | 'terraform'>('types'),
}

const Harness = () => {
  const location = useLocation()
  const list = useListQueryState({ pageSize: 20, filters: FILTERS })

  return (
    <>
      <output data-testid="location">{location.search}</output>
      <output data-testid="search">{list.search}</output>
      <output data-testid="offset">{list.offset}</output>
      <output data-testid="types">
        {[...list.filters.types].sort().join(',')}
      </output>
      <output data-testid="query-key">{JSON.stringify(list.queryKey)}</output>
      <button onClick={() => list.setSearch('payments')}>Search</button>
      <button
        onClick={() => list.setFilter('types', new Set(['terraform', 'helm']))}
      >
        Filter
      </button>
      <button
        onClick={() =>
          list.setFilters({
            active: true,
            types: new Set(['helm']),
          })
        }
      >
        Set filters
      </button>
      <button onClick={() => list.resetFilters()}>Reset filters</button>
      <button onClick={() => list.setOffset(list.offset + list.pageSize)}>
        Next
      </button>
    </>
  )
}

const renderHarness = (initialEntries = ['/items']) => {
  const router = createMemoryRouter([{ path: '*', element: <Harness /> }], {
    initialEntries,
  })
  render(<RouterProvider router={router} />)
  return router
}

test('parses deep-linked state and creates a deterministic query key', () => {
  renderHarness(['/items?types=terraform,helm&q=api&offset=20'])

  expect(screen.getByTestId('search').textContent).toBe('api')
  expect(screen.getByTestId('offset').textContent).toBe('20')
  expect(screen.getByTestId('types').textContent).toBe('helm,terraform')
  expect(screen.getByTestId('query-key').textContent).toBe(
    '["api",20,20,"active",false,"types",["helm","terraform"]]'
  )
})

test('search replaces history, resets offset, and preserves unrelated params', () => {
  const router = renderHarness(['/items?offset=20&panel=details'])

  fireEvent.click(screen.getByRole('button', { name: 'Search' }))

  expect(String(router.state.historyAction)).toBe('REPLACE')
  expect(screen.getByTestId('location').textContent).toBe(
    '?panel=details&q=payments'
  )
})

test('filter updates are atomic and reset offset', () => {
  renderHarness(['/items?offset=40&modal=create'])

  fireEvent.click(screen.getByRole('button', { name: 'Set filters' }))

  const params = new URLSearchParams(
    screen.getByTestId('location').textContent ?? ''
  )
  expect(params.get('active')).toBe('true')
  expect(params.get('types')).toBe('helm')
  expect(params.has('offset')).toBe(false)
  expect(params.get('modal')).toBe('create')
})

test('filter serialization is stable regardless of insertion order', () => {
  renderHarness()

  fireEvent.click(screen.getByRole('button', { name: 'Filter' }))

  expect(screen.getByTestId('location').textContent).toBe(
    '?types=helm%2Cterraform'
  )
})

test('reset removes only owned filter params and offset', () => {
  renderHarness(['/items?types=helm&active=true&offset=20&panel=details&q=api'])

  fireEvent.click(screen.getByRole('button', { name: 'Reset filters' }))

  const params = new URLSearchParams(
    screen.getByTestId('location').textContent ?? ''
  )
  expect(params.get('q')).toBe('api')
  expect(params.get('panel')).toBe('details')
  expect(params.has('types')).toBe(false)
  expect(params.has('active')).toBe(false)
  expect(params.has('offset')).toBe(false)
})

test('pagination pushes history so Back restores the prior page', async () => {
  const router = renderHarness(['/items?q=api'])

  fireEvent.click(screen.getByRole('button', { name: 'Next' }))

  expect(String(router.state.historyAction)).toBe('PUSH')
  expect(screen.getByTestId('location').textContent).toBe('?q=api&offset=20')

  await act(() => router.navigate(-1))

  expect(screen.getByTestId('location').textContent).toBe('?q=api')
  expect(screen.getByTestId('offset').textContent).toBe('0')
})

test('invalid offsets safely resolve to the first page', () => {
  renderHarness(['/items?offset=invalid'])

  expect(screen.getByTestId('offset').textContent).toBe('0')
})
