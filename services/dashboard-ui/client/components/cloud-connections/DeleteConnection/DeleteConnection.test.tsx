import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { DeleteConnection } from './DeleteConnection'

afterEach(cleanup)

test('deletion is blocked while image repositories are assigned', () => {
  for (const count of [0, 1, 3]) {
    let deletes = 0
    render(
      <MemoryRouter>
        <SurfacesProvider>
          <DeleteConnection
            name="acme-images"
            orgId="org-mock-001"
            imageRepositories={count}
            isPending={false}
            onDelete={() => {
              deletes++
            }}
            isVisible
          />
        </SurfacesProvider>
      </MemoryRouter>
    )
    const warning = screen.queryByText(/before deleting it/)
    fireEvent.click(screen.getByRole('button', { name: 'Delete connection' }))
    if (count > 0) {
      expect(warning).toHaveTextContent(
        `assigned to ${count} image ${count === 1 ? 'repository' : 'repositories'}.`
      )
      expect(warning?.querySelector('a')).toHaveTextContent(
        'Image repositories'
      )
      expect(warning?.querySelector('a')).toHaveAttribute(
        'href',
        '/org-mock-001/settings/image-repositories'
      )
      expect(deletes).toBe(0)
    } else {
      expect(warning).toBeNull()
      expect(deletes).toBe(1)
    }
    cleanup()
  }
})
