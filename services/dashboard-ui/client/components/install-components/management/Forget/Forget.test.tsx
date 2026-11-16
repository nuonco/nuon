import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ModalStory } from '@/components/__stories__/helpers'
import { ForgetComponentModal } from './Forget'

const noop = () => {}

afterEach(cleanup)

function setup(appSyncDisabled: boolean) {
  render(
    <MemoryRouter>
      <SurfacesProvider>
        <ModalStory>
          <ForgetComponentModal
            componentName="web-server"
            isLoading={false}
            error={null}
            onConfirm={noop}
            isTornDown
            isInConfig
            appSyncDisabled={appSyncDisabled}
          />
        </ModalStory>
      </SurfacesProvider>
    </MemoryRouter>
  )
  fireEvent.click(screen.getByRole('button', { name: 'Open modal' }))
}

test('flag off tells the user to sync or run the branch', () => {
  setup(false)

  expect(
    screen.getByText(
      (_, element) =>
        element?.textContent ===
        'Remove the component from your config, then run nuon apps sync or run your app branch to update the app configuration.'
    )
  ).toBeTruthy()
})

test('when app sync is disabled tells the user to run their app branch', () => {
  setup(true)

  expect(
    screen.getByText(
      (_, element) =>
        element?.textContent ===
        'Remove the component from your config, then run your app branch to update the app configuration.'
    )
  ).toBeTruthy()
  expect(screen.queryByText('nuon apps sync')).toBeNull()
})
