import { afterEach, expect, mock, test } from 'bun:test'
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import type { TApp, TAppInputConfig } from '@/types'
import { CreateInstallFormFields } from './CreateInstallFormFields'

mock.module('@/hooks/use-surfaces', () => ({
  useSurfaces: () => ({ addModal: () => 'modal-id', removeModal: () => {} }),
}))

const app = { id: 'app-1', name: 'acme' } as TApp
const inputConfig = { id: 'config-1', input_groups: [] } as unknown as TAppInputConfig

afterEach(() => {
  cleanup()
  localStorage.clear()
})

const renderForm = (
  validateName: (name: string) => Promise<string | undefined>,
  onSubmit: (values: unknown) => void = () => {}
) => {
  const states: { canSubmit: boolean; submit: () => unknown }[] = []

  render(
    <CreateInstallFormFields
      app={app}
      inputConfig={inputConfig}
      validateName={validateName}
      onSubmit={onSubmit}
      onStateChange={(state) =>
        states.push({ canSubmit: state.canSubmit, submit: state.submit })
      }
    />
  )

  return states
}

test('surfaces a duplicate name and blocks submit', async () => {
  const states = renderForm(async (name) =>
    name === 'staging' ? 'An install named "staging" already exists' : undefined
  )

  fireEvent.change(screen.getByPlaceholderText('Enter install name'), {
    target: { value: 'staging' },
  })

  await waitFor(() =>
    expect(
      screen.getByText('An install named "staging" already exists')
    ).toBeTruthy()
  )
  expect(states.at(-1)?.canSubmit).toBe(false)
})

test('the name is rechecked on submit, not just while typing', async () => {
  // The name is available while typing and taken by the time submit runs, which
  // is only caught if submit revalidates rather than trusting the change check.
  const taken = new Set<string>()
  const submitted: unknown[] = []
  const states = renderForm(
    async (name) =>
      taken.has(name) ? `An install named "${name}" already exists` : undefined,
    (values) => submitted.push(values)
  )

  fireEvent.change(screen.getByPlaceholderText('Enter install name'), {
    target: { value: 'staging' },
  })
  await waitFor(() => expect(states.at(-1)?.canSubmit).toBe(true))

  taken.add('staging')
  await act(async () => {
    await states.at(-1)!.submit()
  })

  await waitFor(() =>
    expect(
      screen.getByText('An install named "staging" already exists')
    ).toBeTruthy()
  )
  expect(submitted).toHaveLength(0)
})

test('an available name leaves the form submittable', async () => {
  const states = renderForm(async () => undefined)

  fireEvent.change(screen.getByPlaceholderText('Enter install name'), {
    target: { value: 'fresh-name' },
  })

  await waitFor(() => expect(states.at(-1)?.canSubmit).toBe(true))
  expect(screen.queryByText(/already exists/)).toBeNull()
})
