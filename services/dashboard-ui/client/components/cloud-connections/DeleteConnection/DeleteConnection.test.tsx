import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { DeleteConnection } from './DeleteConnection'

afterEach(cleanup)

test('deletion warns about pinned image configurations without blocking intentional removal', () => {
  for (const count of [0, 1, 3]) {
    render(
      <MemoryRouter>
        <SurfacesProvider>
          <DeleteConnection
            name="acme-images"
            imageConfigs={count}
            isPending={false}
            onDelete={() => {}}
            isVisible
          />
        </SurfacesProvider>
      </MemoryRouter>
    )
    const warning = screen.queryByText(/will cause future builds/)
    if (count > 0) {
      expect(warning).toHaveTextContent(
        `${count} saved image ${count === 1 ? 'configuration.' : 'configurations.'}`
      )
      expect(warning).toHaveTextContent('metadata fetches')
      expect(warning).toHaveTextContent('re-sync the affected apps')
    } else {
      expect(warning).toBeNull()
    }
    expect(
      screen.getByRole('button', { name: 'Delete connection' })
    ).toBeEnabled()
    cleanup()
  }
})
