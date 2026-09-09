import { afterEach, expect, mock, test } from 'bun:test'
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { Banner } from './Banner'

const originalMatchMedia = window.matchMedia

const installMatchMedia = (matches: boolean) => {
  window.matchMedia = mock(
    (query: string) =>
      ({
        matches,
        media: query,
        onchange: null,
        addEventListener: () => {},
        removeEventListener: () => {},
        addListener: () => {},
        removeListener: () => {},
        dispatchEvent: () => true,
      }) as MediaQueryList
  )
}

afterEach(() => {
  cleanup()
  window.matchMedia = originalMatchMedia
})

test('renders heading and body text', () => {
  render(
    <Banner heading="Deploy failed">
      Unable to deploy the component.
    </Banner>
  )

  expect(screen.getByText('Deploy failed')).toBeTruthy()
  expect(screen.getByText('Unable to deploy the component.')).toBeTruthy()
})

test('assigns urgency from every theme', () => {
  const themes = [
    ['error', 'alert', 'assertive'],
    ['warn', 'alert', 'assertive'],
    ['default', 'status', 'polite'],
    ['success', 'status', 'polite'],
    ['info', 'status', 'polite'],
    ['brand', 'status', 'polite'],
    ['neutral', 'status', 'polite'],
  ] as const

  for (const [theme, role, live] of themes) {
    const { unmount } = render(<Banner theme={theme} heading={theme} />)
    const banner = screen.getByRole(role)

    expect(banner.getAttribute('aria-live')).toBe(live)
    expect(banner.getAttribute('data-banner-theme')).toBe(theme)
    unmount()
  }
})

test('only renders a dismiss control when dismissible', () => {
  const { rerender } = render(<Banner heading="Status update" />)

  expect(screen.queryByRole('button', { name: 'Dismiss' })).toBeNull()

  rerender(<Banner heading="Status update" dismissible />)

  expect(screen.getByRole('button', { name: 'Dismiss' })).toBeTruthy()
})

test('uses the custom dismiss label', () => {
  render(
    <Banner heading="Status update" dismissible dismissLabel="Close message" />
  )

  expect(screen.getByRole('button', { name: 'Close message' })).toBeTruthy()
})

test('removes itself and calls onDismiss once', async () => {
  installMatchMedia(false)
  const onDismiss = mock(() => {})
  render(
    <Banner heading="Status update" dismissible onDismiss={onDismiss} />
  )

  fireEvent.click(screen.getByRole('button', { name: 'Dismiss' }))

  await waitFor(() => {
    expect(onDismiss).toHaveBeenCalledTimes(1)
  })
  expect(screen.queryByText('Status update')).toBeNull()
})

test('dismisses without an onDismiss callback', async () => {
  installMatchMedia(true)
  render(<Banner heading="Status update" dismissible />)

  fireEvent.click(screen.getByRole('button', { name: 'Dismiss' }))

  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 250))
  })
  await waitFor(() => {
    expect(screen.queryByText('Status update')).toBeNull()
  })
})

test('reduced motion removes the banner without the exit delay', async () => {
  installMatchMedia(true)
  const onDismiss = mock(() => {})
  render(
    <Banner heading="Status update" dismissible onDismiss={onDismiss} />
  )

  fireEvent.click(screen.getByRole('button', { name: 'Dismiss' }))

  await waitFor(() => {
    expect(onDismiss).toHaveBeenCalledTimes(1)
  })
  expect(screen.queryByText('Status update')).toBeNull()
})

test('does not call onDismiss after unmounting during exit', async () => {
  installMatchMedia(false)
  const onDismiss = mock(() => {})
  const { unmount } = render(
    <Banner heading="Status update" dismissible onDismiss={onDismiss} />
  )

  fireEvent.click(screen.getByRole('button', { name: 'Dismiss' }))
  unmount()
  await new Promise((resolve) => setTimeout(resolve, 250))

  expect(onDismiss).not.toHaveBeenCalled()
})

test('lets a caller override the default role', () => {
  render(<Banner theme="error" role="note" heading="Status update" />)

  expect(screen.getByRole('note')).toBeTruthy()
  expect(screen.queryByRole('alert')).toBeNull()
})
