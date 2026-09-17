import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { StatusBarProvider } from '../providers/status-bar-provider'
import { useStatusBar, useStatusBarContent } from './use-status-bar'

afterEach(cleanup)

const Bar = () => (
  <span data-testid="status-bar">{useStatusBarContent() ?? 'org only'}</span>
)

const Child = ({ label }: { label: string }) => {
  useStatusBar(<span>{label}</span>, [label])
  return null
}

test('clears the registered context when its route unmounts', async () => {
  const view = render(
    <StatusBarProvider>
      <Bar />
      <Child label="payments / main" />
    </StatusBarProvider>
  )

  await waitFor(() =>
    expect(screen.getByTestId('status-bar').textContent).toBe('payments / main')
  )

  view.rerender(
    <StatusBarProvider>
      <Bar />
    </StatusBarProvider>
  )

  await waitFor(() =>
    expect(screen.getByTestId('status-bar').textContent).toBe('org only')
  )
})

test('replaces the context when another route registers', async () => {
  const view = render(
    <StatusBarProvider>
      <Bar />
      <Child label="payments / main" />
    </StatusBarProvider>
  )

  await waitFor(() =>
    expect(screen.getByTestId('status-bar').textContent).toBe('payments / main')
  )

  view.rerender(
    <StatusBarProvider>
      <Bar />
      <Child label="payments / release" />
    </StatusBarProvider>
  )

  await waitFor(() =>
    expect(screen.getByTestId('status-bar').textContent).toBe(
      'payments / release'
    )
  )
})
