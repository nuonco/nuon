import { afterEach, beforeEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router'
import { useSearchParamState } from './use-search-param-state'
import { useReplaceSearchParams } from './use-replace-search-params'

const setURL = (search = '') => {
  ;(window as any).happyDOM.setURL(`http://localhost/page${search}`)
}

beforeEach(() => setURL('?panel=group-1'))
afterEach(() => {
  cleanup()
  ;(window as any).happyDOM.setURL('about:blank')
})

const Probe = () => {
  const replaceSearch = useReplaceSearchParams()
  const [, setTab] = useSearchParamState('step_tab')
  return (
    <>
      <button onClick={() => replaceSearch((params) => params.set('view', 'raw'))}>
        Set view
      </button>
      <button onClick={() => setTab('logs')}>Set tab</button>
      <button
        onClick={() => {
          replaceSearch((params) => params.set('view', 'raw'))
          setURL('?panel=group-1')
          setTab('logs')
        }}
      >
        Set both while the bar lags
      </button>
    </>
  )
}

test('a later writer keeps params an earlier writer just set', () => {
  render(
    <BrowserRouter>
      <Probe />
    </BrowserRouter>
  )

  fireEvent.click(screen.getByText('Set view'))
  fireEvent.click(screen.getByText('Set tab'))

  const params = new URLSearchParams(window.location.search)
  expect(params.getAll('panel')).toEqual(['group-1'])
  expect(params.get('view')).toBe('raw')
  expect(params.get('step_tab')).toBe('logs')
})

test('a later writer keeps a param the address bar does not show yet', () => {
  render(
    <BrowserRouter>
      <Probe />
    </BrowserRouter>
  )

  fireEvent.click(screen.getByText('Set both while the bar lags'))

  const params = new URLSearchParams(window.location.search)
  expect(params.get('view')).toBe('raw')
  expect(params.get('step_tab')).toBe('logs')
  expect(params.getAll('panel')).toEqual(['group-1'])
})
