import { afterEach, expect, jest, mock, test } from 'bun:test'
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { useToast } from '../../../hooks/use-toast'
import {
  DEFAULT_TOAST_TIMEOUT,
  ToastProvider,
} from '../../../providers/toast-provider'
import { Button } from '../../atoms/Button'
import { TOAST_EXIT_MS } from './toast-motion'

afterEach(() => {
  jest.useRealTimers()
  cleanup()
})

const ToastControls = ({ onAction }: { onAction?: () => void }) => {
  const { addToast, clearToasts } = useToast()
  return (
    <>
      <Button
        onClick={() =>
          addToast({
            heading: 'Deploy started',
            description: 'Deploying payments to production.',
            theme: 'info',
            timeout: null,
          })
        }
      >
        Add toast
      </Button>
      <Button
        onClick={() => {
          for (let index = 1; index <= 6; index += 1) {
            addToast({
              heading: `Notification ${index}`,
              theme: index === 1 ? 'error' : 'neutral',
              timeout: null,
            })
          }
        }}
      >
        Add six
      </Button>
      <Button
        onClick={() =>
          addToast({
            heading: 'Plan ready',
            action: { label: 'View plan', onClick: onAction ?? (() => {}) },
            timeout: null,
          })
        }
      >
        Add action
      </Button>
      <Button
        onClick={() =>
          addToast({
            heading: 'Timed toast',
            timeout: 1000,
          })
        }
      >
        Add timed
      </Button>
      <Button onClick={clearToasts}>Clear</Button>
    </>
  )
}

const renderToasts = (onAction?: () => void) =>
  render(
    <ToastProvider>
      <ToastControls onAction={onAction} />
    </ToastProvider>
  )

test('adds and manually dismisses a toast through one portal', async () => {
  jest.useFakeTimers()
  renderToasts()

  fireEvent.click(screen.getByRole('button', { name: 'Add toast' }))

  expect(screen.getByRole('status')).toHaveTextContent('Deploy started')
  expect(document.querySelectorAll('#lite-toast-root')).toHaveLength(1)

  fireEvent.click(
    screen.getByRole('button', {
      name: 'Dismiss Deploy started notification',
    })
  )
  expect(
    document.querySelector<HTMLElement>('[data-toast-slot]')?.dataset.state
  ).toBe('exiting')

  act(() => {
    jest.advanceTimersByTime(TOAST_EXIT_MS + 40)
  })
  expect(screen.queryByText('Deploy started')).toBeNull()
})

test('keeps burst additions and expands every active toast', async () => {
  renderToasts()

  fireEvent.click(screen.getByRole('button', { name: 'Add six' }))

  await waitFor(() => {
    expect(document.querySelectorAll('[data-toast-slot]')).toHaveLength(6)
  })

  const stack = document.querySelector<HTMLElement>('[data-toast-stack]')!
  const collapsed = Array.from(
    document.querySelectorAll<HTMLElement>('[data-toast-slot]')
  )
  expect(
    collapsed.filter((slot) => slot.getAttribute('aria-hidden') === 'true')
  ).toHaveLength(5)

  fireEvent.mouseEnter(stack)

  await waitFor(() => {
    expect(stack.getAttribute('data-expanded')).toBe('true')
    expect(
      Array.from(
        document.querySelectorAll<HTMLElement>('[data-toast-slot]')
      ).every((slot) => !slot.hasAttribute('aria-hidden'))
    ).toBe(true)
  })
})

test('collapses after removing a focused toast from the middle', async () => {
  jest.useFakeTimers()
  renderToasts()

  fireEvent.click(screen.getByRole('button', { name: 'Add six' }))
  const stack = document.querySelector<HTMLElement>('[data-toast-stack]')!
  fireEvent.mouseEnter(stack)

  const dismiss = screen.getByRole('button', {
    name: 'Dismiss Notification 3 notification',
  })
  fireEvent.focus(dismiss)
  fireEvent.click(dismiss)
  fireEvent.mouseLeave(stack)

  expect(stack.hasAttribute('data-expanded')).toBe(false)

  act(() => {
    jest.advanceTimersByTime(TOAST_EXIT_MS + 40)
  })

  expect(stack.hasAttribute('data-expanded')).toBe(false)
})

test('uses assertive alerts for warning and error themes', async () => {
  renderToasts()

  fireEvent.click(screen.getByRole('button', { name: 'Add six' }))
  const stack = document.querySelector<HTMLElement>('[data-toast-stack]')!
  fireEvent.mouseEnter(stack)

  const alert = await screen.findByRole('alert')
  expect(alert.getAttribute('aria-live')).toBe('assertive')
  expect(alert.getAttribute('aria-atomic')).toBe('true')
})

test('runs one optional action and dismisses its toast', async () => {
  jest.useFakeTimers()
  const onAction = mock(() => {})
  renderToasts(onAction)

  fireEvent.click(screen.getByRole('button', { name: 'Add action' }))
  fireEvent.click(screen.getByRole('button', { name: 'View plan' }))

  expect(onAction).toHaveBeenCalledTimes(1)
  act(() => {
    jest.advanceTimersByTime(TOAST_EXIT_MS + 40)
  })
  expect(screen.queryByText('Plan ready')).toBeNull()
})

test('pauses and resumes the remaining dismissal duration', async () => {
  jest.useFakeTimers()
  renderToasts()

  fireEvent.click(screen.getByRole('button', { name: 'Add timed' }))
  const stack = document.querySelector<HTMLElement>('[data-toast-stack]')!

  act(() => jest.advanceTimersByTime(400))
  fireEvent.mouseEnter(stack)
  act(() => jest.advanceTimersByTime(DEFAULT_TOAST_TIMEOUT))
  expect(screen.getByText('Timed toast')).toBeTruthy()

  fireEvent.mouseLeave(stack)
  act(() => jest.advanceTimersByTime(599))
  expect(screen.getByText('Timed toast')).toBeTruthy()

  act(() => jest.advanceTimersByTime(1))
  act(() => jest.advanceTimersByTime(TOAST_EXIT_MS + 40))
  expect(screen.queryByText('Timed toast')).toBeNull()
})

test('clears every active toast', async () => {
  jest.useFakeTimers()
  renderToasts()

  fireEvent.click(screen.getByRole('button', { name: 'Add six' }))
  expect(document.querySelectorAll('[data-toast-slot]')).toHaveLength(6)

  fireEvent.click(screen.getByRole('button', { name: 'Clear' }))

  act(() => {
    jest.advanceTimersByTime(TOAST_EXIT_MS + 40)
  })
  expect(document.querySelectorAll('[data-toast-slot]')).toHaveLength(0)
})
