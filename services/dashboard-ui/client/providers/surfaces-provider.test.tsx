import { afterEach, beforeEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter, useSearchParams } from 'react-router'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { SurfacesProvider } from './surfaces-provider'

const setURL = (search = '') => {
  ;(window as any).happyDOM.setURL(`http://localhost/page${search}`)
}

beforeEach(() => setURL())
afterEach(() => {
  cleanup()
  ;(window as any).happyDOM.setURL('about:blank')
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
        onClick={() =>
          addPanel(
            <Panel heading="Group">
              <button
                onClick={() =>
                  addPanel(<Panel heading="Plan diff" />, 'plan-diff-1')
                }
              >
                Open nested plan diff
              </button>
            </Panel>,
            'group-1'
          )
        }
      >
        Open group with trigger
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
    <BrowserRouter>
      <SurfacesProvider>
        <KeyedProbe />
      </SurfacesProvider>
    </BrowserRouter>
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

test('a panel opened from inside another panel keeps the parent key', () => {
  renderKeyedProbe()

  fireEvent.click(screen.getByText('Open group with trigger'))
  fireEvent.click(screen.getByText('Open nested plan diff'))

  expect(screen.getByTestId('search').textContent).toBe(
    'panel=group-1&panel=plan-diff-1'
  )
})

test('deep-linked panels stack in url order when opened out of order', () => {
  setURL('?panel=group-1&panel=plan-diff-1')
  render(
    <BrowserRouter>
      <SurfacesProvider>
        <KeyedProbe />
      </SurfacesProvider>
    </BrowserRouter>
  )

  fireEvent.click(screen.getByText('Open plan diff'))
  fireEvent.click(screen.getByText('Open group'))

  const text = document.getElementById('panel-root')?.textContent ?? ''
  expect(screen.getByTestId('order').textContent).toBe('group-1,plan-diff-1')
  expect(text.indexOf('Group')).toBeLessThan(text.indexOf('Plan diff'))
})
