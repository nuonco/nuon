import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, useSearchParams } from 'react-router'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { SurfacesProvider } from './surfaces-provider'

afterEach(cleanup)

const Probe = () => {
  const { addPanel, addPanelIfEmpty, panels } = useSurfaces()
  return (
    <>
      <div data-testid="panel-count">{panels.length}</div>
      <button onClick={() => addPanel(<Panel heading="Manual" />)}>
        Open manual panel
      </button>
      <button
        onClick={() => {
          addPanelIfEmpty(<Panel heading="First automatic" />)
          addPanelIfEmpty(<Panel heading="Second automatic" />)
        }}
      >
        Auto-open panels
      </button>
    </>
  )
}

const renderProbe = () =>
  render(
    <MemoryRouter>
      <SurfacesProvider>
        <Probe />
      </SurfacesProvider>
    </MemoryRouter>
  )

test('automatic panels do not open over an existing panel', () => {
  renderProbe()

  fireEvent.click(screen.getByText('Open manual panel'))
  fireEvent.click(screen.getByText('Auto-open panels'))

  expect(screen.getByTestId('panel-count')).toHaveTextContent('1')
  expect(screen.getByText('Manual')).toBeTruthy()
  expect(screen.queryByText('First automatic')).toBeNull()
  expect(screen.queryByText('Second automatic')).toBeNull()
})

test('simultaneous automatic panels open at most one panel', () => {
  renderProbe()

  fireEvent.click(screen.getByText('Auto-open panels'))

  expect(screen.getByTestId('panel-count')).toHaveTextContent('1')
  expect(screen.getByText('First automatic')).toBeTruthy()
  expect(screen.queryByText('Second automatic')).toBeNull()
})

const KeyedProbe = () => {
  const { addPanel, removePanel, panels } = useSurfaces()
  const [searchParams] = useSearchParams()
  return (
    <>
      <div data-testid="search">{searchParams.toString()}</div>
      <div data-testid="order">{panels.map((panel) => panel.key).join(',')}</div>
      <button onClick={() => addPanel(<Panel heading="Group" />, 'group-1')}>
        Open group
      </button>
      <button
        onClick={() =>
          addPanel(<Panel heading="Plan diff" />, 'plan-diff-1')
        }
      >
        Open plan diff
      </button>
      <button
        onClick={() => {
          const top = panels.at(-1)
          if (top) removePanel(top.id, top.key)
        }}
      >
        Close top panel
      </button>
    </>
  )
}

const renderKeyedProbe = () =>
  render(
    <MemoryRouter>
      <SurfacesProvider>
        <KeyedProbe />
      </SurfacesProvider>
    </MemoryRouter>
  )

test('stacking a keyed panel over another keyed panel keeps both keys in the url', async () => {
  renderKeyedProbe()

  fireEvent.click(screen.getByText('Open group'))
  expect(screen.getByTestId('search')).toHaveTextContent('panel=group-1')

  fireEvent.click(screen.getByText('Open plan diff'))
  expect(screen.getByTestId('search').textContent).toContain('group-1')
  expect(screen.getByTestId('search').textContent).toContain('plan-diff-1')

  fireEvent.click(screen.getByText('Close top panel'))
  await waitFor(() =>
    expect(screen.getByTestId('search').textContent).not.toContain(
      'plan-diff-1'
    )
  )
  expect(screen.getByTestId('search').textContent).toContain('group-1')
})

test('deep-linked panels stack in url order when opened out of order', () => {
  render(
    <MemoryRouter initialEntries={['/?panel=group-1&panel=plan-diff-1']}>
      <SurfacesProvider>
        <KeyedProbe />
      </SurfacesProvider>
    </MemoryRouter>
  )

  fireEvent.click(screen.getByText('Open plan diff'))
  fireEvent.click(screen.getByText('Open group'))

  const text = document.getElementById('panel-root')?.textContent ?? ''
  expect(screen.getByTestId('order').textContent).toBe('group-1,plan-diff-1')
  expect(text.indexOf('Group')).toBeLessThan(text.indexOf('Plan diff'))
})
