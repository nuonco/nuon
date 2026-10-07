import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { DRAFT_VERSION } from '@/hooks/use-draft-persistence'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TApp, TAppInputConfig } from '@/types'
import { CreateInstallFormFields } from './CreateInstallFormFields'

const app = {
  id: 'app-1',
  runner_config: { app_runner_type: 'aws' },
} as TApp

const inputConfig = {
  id: 'cfg-1',
  input_groups: [],
  inputs: [],
} as TAppInputConfig

afterEach(() => {
  cleanup()
  localStorage.clear()
})

test('resuming a draft fills the install form', async () => {
  localStorage.setItem(
    'install-draft:app-1',
    JSON.stringify({
      values: {
        name: 'kept-name',
        region: 'us-west-2',
        inputs: {},
      },
      timestamp: new Date().toISOString(),
      version: DRAFT_VERSION,
      configId: 'cfg-1',
    })
  )

  let canSubmit = false
  render(
    <MemoryRouter>
      <SurfacesProvider>
        <CreateInstallFormFields
          app={app}
          inputConfig={inputConfig}
          onSubmit={() => {}}
          onStateChange={(state) => {
            canSubmit = state.canSubmit
          }}
        />
      </SurfacesProvider>
    </MemoryRouter>
  )

  fireEvent.click(await screen.findByRole('button', { name: 'Resume draft' }))

  const name = screen.getByPlaceholderText('Enter install name')
  expect(name).toHaveValue('kept-name')
  fireEvent.focus(name)
  fireEvent.blur(name)
  expect(screen.queryByText('Install name is required')).toBeNull()
  await waitFor(() => expect(canSubmit).toBe(true))
})
