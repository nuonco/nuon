import { afterEach, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import type { TApp, TAppInputConfig } from '@/types'
import { CreateInstallFormFields } from './CreateInstallFormFields'

mock.module('@/hooks/use-surfaces', () => ({
  useSurfaces: () => ({ addModal: () => 'modal-id', removeModal: () => {} }),
}))

const app = { id: 'app-1', name: 'acme' } as TApp
const inputConfig = { id: 'config-1', input_groups: [] } as unknown as TAppInputConfig

afterEach(cleanup)

const renderForm = (validateName: (name: string) => Promise<string | undefined>) => {
  const states: { canSubmit: boolean }[] = []

  render(
    <CreateInstallFormFields
      app={app}
      inputConfig={inputConfig}
      validateName={validateName}
      onSubmit={() => {}}
      onStateChange={(state) => states.push({ canSubmit: state.canSubmit })}
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

test('an available name leaves the form submittable', async () => {
  const states = renderForm(async () => undefined)

  fireEvent.change(screen.getByPlaceholderText('Enter install name'), {
    target: { value: 'fresh-name' },
  })

  await waitFor(() => expect(states.at(-1)?.canSubmit).toBe(true))
  expect(screen.queryByText(/already exists/)).toBeNull()
})
