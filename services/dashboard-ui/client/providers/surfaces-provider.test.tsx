import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
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
