import { afterEach, expect, test } from 'bun:test'
import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import { useState, type ReactNode } from 'react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { useWizardHistory } from './use-wizard-history'

afterEach(cleanup)

function Steps({
  initial,
  onExit,
}: {
  initial: number
  onExit?: () => void
}) {
  const [step, setStep] = useState(initial)
  useWizardHistory(step, () => setStep((current) => Math.max(0, current - 1)), onExit)
  return <p>step {step}</p>
}

function renderAt(element: ReactNode) {
  const router = createMemoryRouter(
    [
      { path: '/', element: <p>home</p> },
      { path: '/onboarding', element },
    ],
    { initialEntries: ['/', '/onboarding'], initialIndex: 1 }
  )
  render(<RouterProvider router={router} />)
  return router
}

test('browser back moves to the previous wizard step, then leaves', async () => {
  const router = renderAt(<Steps initial={2} />)
  expect(screen.getByText('step 2')).toBeTruthy()

  await act(() => router.navigate(-1))
  await waitFor(() => expect(screen.getByText('step 1')).toBeTruthy())

  await act(() => router.navigate(-1))
  await waitFor(() => expect(screen.getByText('step 0')).toBeTruthy())

  await act(() => router.navigate(-1))
  await waitFor(() => expect(screen.getByText('home')).toBeTruthy())
})

test('browser back from the first step can return to the intro', async () => {
  function Flow() {
    const [intro, setIntro] = useState(false)
    if (intro) return <p>intro</p>
    return <Steps initial={0} onExit={() => setIntro(true)} />
  }

  const router = renderAt(<Flow />)
  await act(() => router.navigate(-1))
  await waitFor(() => expect(screen.getByText('intro')).toBeTruthy())

  await act(() => router.navigate(-1))
  await waitFor(() => expect(screen.getByText('home')).toBeTruthy())
})
