import { afterEach, beforeEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { useState } from 'react'
import { BrowserRouter, useSearchParams } from 'react-router'
import { useClearSearchParamsOnUnmount } from './use-clear-search-params-on-unmount'

beforeEach(() => {
  ;(window as any).happyDOM.setURL('http://localhost/page?view=raw&keep=1')
})
afterEach(() => {
  cleanup()
  ;(window as any).happyDOM.setURL('about:blank')
})

const Owner = () => {
  useClearSearchParamsOnUnmount(['view'])
  return null
}

const Harness = ({ owners = 1 }: { owners?: number }) => {
  const [mounted, setMounted] = useState(owners)
  const [searchParams] = useSearchParams()
  return (
    <>
      <div data-testid="search">{searchParams.toString()}</div>
      <button onClick={() => setMounted((n) => n - 1)}>Unmount one</button>
      {Array.from({ length: mounted }, (_, i) => (
        <Owner key={i} />
      ))}
    </>
  )
}

test('clears owned params when the owner unmounts', async () => {
  render(
    <BrowserRouter>
      <Harness />
    </BrowserRouter>
  )
  fireEvent.click(screen.getByText('Unmount one'))
  await waitFor(() =>
    expect(screen.getByTestId('search').textContent).toBe('keep=1')
  )
})

test('keeps params while another owner is still mounted', async () => {
  render(
    <BrowserRouter>
      <Harness owners={2} />
    </BrowserRouter>
  )
  fireEvent.click(screen.getByText('Unmount one'))
  await new Promise((resolve) => setTimeout(resolve, 20))
  expect(screen.getByTestId('search').textContent).toBe('view=raw&keep=1')
})

test('leaves the url alone after navigating to another page', async () => {
  const { unmount } = render(
    <BrowserRouter>
      <Owner />
    </BrowserRouter>
  )
  window.history.replaceState(null, '', '/other?view=raw')
  unmount()
  await new Promise((resolve) => setTimeout(resolve, 20))
  expect(window.location.search).toBe('?view=raw')
})
