import { afterEach, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { installsListQuery } from '../queries/installs'
import { useRoutePrefetch } from './use-route-prefetch'

afterEach(cleanup)

const setup = () => {
  const queryClient = new QueryClient()
  const prefetchQuery = mock(() => Promise.resolve())
  queryClient.prefetchQuery = prefetchQuery as never

  const Harness = () => {
    useRoutePrefetch()
    return (
      <>
        <a href="/org-123/installs">
          <span>Installs</span>
        </a>
        <a href="/org-123/apps?q=payments">Filtered apps</a>
        <a href="https://docs.nuon.co">Docs</a>
      </>
    )
  }

  render(
    <QueryClientProvider client={queryClient}>
      <Harness />
    </QueryClientProvider>
  )

  return { prefetchQuery }
}

const keysPrefetched = (prefetchQuery: ReturnType<typeof mock>) =>
  prefetchQuery.mock.calls.map(
    ([options]: [{ queryKey: unknown[] }]) => options.queryKey
  )

test('hovering an internal link warms the destination list', () => {
  const { prefetchQuery } = setup()

  fireEvent.pointerOver(screen.getByText('Installs'))

  expect(keysPrefetched(prefetchQuery)).toEqual([
    installsListQuery({ orgId: 'org-123' }).queryKey,
  ])
})

test('focusing an internal link warms the destination list', () => {
  const { prefetchQuery } = setup()

  fireEvent.focusIn(screen.getByRole('link', { name: 'Installs' }))

  expect(prefetchQuery).toHaveBeenCalledTimes(1)
})

test('moving across the same link warms it once', () => {
  const { prefetchQuery } = setup()
  const link = screen.getByRole('link', { name: 'Installs' })

  fireEvent.pointerOver(link)
  fireEvent.pointerOver(screen.getByText('Installs'))
  fireEvent.pointerOver(link)

  expect(prefetchQuery).toHaveBeenCalledTimes(1)
})

test('leaves query-scoped and external links alone', () => {
  const { prefetchQuery } = setup()

  fireEvent.pointerOver(screen.getByRole('link', { name: 'Filtered apps' }))
  fireEvent.pointerOver(screen.getByRole('link', { name: 'Docs' }))

  expect(prefetchQuery).not.toHaveBeenCalled()
})
