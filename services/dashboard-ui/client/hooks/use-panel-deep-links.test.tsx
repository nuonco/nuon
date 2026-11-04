import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { usePanelDeepLinks, type TPanelResolver } from './use-panel-deep-links'

afterEach(() => {
  cleanup()
  ;(window as any).happyDOM.setURL('about:blank')
})

const Probe = ({ resolve }: { resolve: TPanelResolver }) => {
  usePanelDeepLinks(resolve)
  const { panels } = useSurfaces()
  return <div data-testid="order">{panels.map((p) => p.key).join(',')}</div>
}

const renderAt = (search: string, resolve: TPanelResolver) => {
  ;(window as any).happyDOM.setURL(`http://localhost/page${search}`)
  return render(
    <BrowserRouter>
      <SurfacesProvider>
        <Probe resolve={resolve} />
      </SurfacesProvider>
    </BrowserRouter>
  )
}

test('opens every resolvable key in url stack order', async () => {
  renderAt('?panel=parent-1&panel=child-1', (key) => (
    <Panel heading={key} />
  ))
  await waitFor(() =>
    expect(screen.getByTestId('order').textContent).toBe('parent-1,child-1')
  )
})

test('skips keys the resolver does not own', async () => {
  renderAt('?panel=other-1&panel=child-1', (key) =>
    key.startsWith('child-') ? <Panel heading={key} /> : null
  )
  await waitFor(() =>
    expect(screen.getByTestId('order').textContent).toBe('child-1')
  )
})

test('opens a key once even across re-renders', async () => {
  let calls = 0
  renderAt('?panel=child-1', (key) => {
    calls += 1
    return <Panel heading={key} />
  })
  await waitFor(() =>
    expect(screen.getByTestId('order').textContent).toBe('child-1')
  )
  expect(calls).toBe(1)
})
