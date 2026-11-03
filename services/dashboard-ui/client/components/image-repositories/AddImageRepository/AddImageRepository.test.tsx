import { expect, test } from 'bun:test'
import { act, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import {
  verifiedECRConnection,
  verifiedGARConnection,
  verifiedACRConnection,
} from '@/components/cloud-connections/__stories__/fixtures'
import { AddImageRepository } from './AddImageRepository'

for (const { identity, name, canonical } of [
  {
    identity: verifiedECRConnection,
    name: 'acme/api',
    canonical: 'arn:aws:ecr:us-west-2:123456789012:repository/acme/api',
  },
  {
    identity: verifiedGARConnection,
    name: 'acme-services',
    canonical:
      'projects/acme-project/locations/us-central1/repositories/acme-services',
  },
  { identity: verifiedACRConnection, name: 'acme/api', canonical: 'acme/api' },
]) {
  test(`${identity.preset} saves the user's explicit identity, not the first matching identity`, async () => {
    let saved: unknown
    const alternative = { ...identity, id: 'cc-second', name: 'acme-second' }
    render(
      <MemoryRouter>
        <AddImageRepository
          orgId="org-mock-001"
          connections={[identity, alternative]}
          initialValues={{
            platform: identity.platform,
            repository: name,
            connection_id: '',
          }}
          onSave={(value) => {
            saved = value
          }}
          onCreateIdentity={async () => alternative}
        />
      </MemoryRouter>
    )
    const save = screen.getByText('Save repository', {
      selector: 'button',
    })
    expect(save).toHaveAttribute('aria-disabled', 'true')
    expect(saved).toBeUndefined()
    fireEvent.click(screen.getByLabelText('Pull identity'))
    await act(async () => {
      fireEvent.click(screen.getByText('acme-second'))
    })
    const selectedSave = screen.getByText('Save repository', {
      selector: 'button',
    })
    expect(selectedSave).not.toHaveAttribute('aria-disabled', 'true')
    expect(selectedSave).toBeEnabled()
    await act(async () => {
      fireEvent.click(selectedSave)
    })
    expect(saved).toEqual({
      cloud_connection_id: 'cc-second',
      repository: canonical,
    })
  })
}

test('the inline identity defaults to a unique generated role name', async () => {
  render(
    <MemoryRouter>
      <AddImageRepository
        orgId="org-mock-001"
        connections={[verifiedECRConnection]}
        onCreateIdentity={async () => verifiedECRConnection}
        onSave={() => {}}
      />
    </MemoryRouter>
  )
  await act(async () => {
    fireEvent.click(
      screen.getByRole('radio', { name: 'Create a new principal' })
    )
  })
  const roleName = screen.getByLabelText('IAM role name') as HTMLInputElement
  expect(roleName.value).toMatch(/^nuon-ecr-pull-[a-z0-9]{4}$/)
})

test('defaults to the existing-identity flow when identities exist for the platform', () => {
  render(
    <MemoryRouter>
      <AddImageRepository
        orgId="org-mock-001"
        connections={[verifiedECRConnection]}
        onCreateIdentity={async () => verifiedECRConnection}
        onSave={() => {}}
      />
    </MemoryRouter>
  )
  expect(
    (
      screen.getByRole('radio', {
        name: 'Use an existing principal',
      }) as HTMLInputElement
    ).checked
  ).toBe(true)
  expect(screen.getByLabelText('Pull identity')).toBeVisible()
  expect(screen.getByLabelText('Registry provider')).toBeVisible()
  expect(screen.getByLabelText('Repository name')).toBeVisible()
})

test('skips the mode choice and shows the create form when no identities exist', () => {
  render(
    <MemoryRouter>
      <AddImageRepository
        orgId="org-mock-001"
        connections={[]}
        onCreateIdentity={async () => verifiedECRConnection}
        onSave={() => {}}
      />
    </MemoryRouter>
  )
  expect(
    screen.queryByRole('radio', { name: 'Create a new principal' })
  ).toBeNull()
  expect(screen.queryByLabelText('Pull identity')).toBeNull()
  expect(screen.getByLabelText('Registry provider')).toBeVisible()
  expect(screen.getByLabelText('Repository name')).toBeVisible()
  const roleName = screen.getByLabelText('IAM role name') as HTMLInputElement
  expect(roleName.value).toMatch(/^nuon-ecr-pull-[a-z0-9]{4}$/)
})
