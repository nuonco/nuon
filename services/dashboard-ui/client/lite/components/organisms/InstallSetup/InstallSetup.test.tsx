import { afterEach, describe, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { loadDraft, saveDraft } from '../../../utils/draft'
import { InstallSetup } from './InstallSetup'

const orgId = 'org_example'
const wizard = 'install-setup-test'

afterEach(() => {
  cleanup()
  localStorage.clear()
})

describe('InstallSetup', () => {
  test('restores draft state, shows group labels, and edits a duplicate name', () => {
    saveDraft(orgId, wizard, {
      progress: { app: true, details: true },
      values: {
        appId: 'app_payments',
        name: 'production',
        location: 'us-west-2',
        accountId: '',
        autoApprove: false,
        stackOnly: false,
        labels: [{ key: 'team', value: 'payments' }],
        inputs: [],
        branchId: 'branch_main',
        installGroupId: 'group_production',
      },
    })
    const onAppChange = mock()
    const onBranchChange = mock()
    const onClearError = mock()

    render(
      <InstallSetup
        apps={[
          {
            id: 'app_payments',
            name: 'Payments API',
            platform: 'aws',
            updatedLabel: 'synced just now',
          },
        ]}
        branches={[
          {
            id: 'branch_main',
            name: 'main',
            groups: [
              {
                id: 'group_production',
                name: 'Production',
                kind: 'labels',
                labels: { env: 'production' },
              },
            ],
          },
        ]}
        inputs={[]}
        platform="aws"
        configurationKey="cfg_active"
        configurationReady
        error={{
          error: 'unable to create install: duplicated key not allowed',
          description: 'duplicate key',
          user_error: true,
          status: 409,
        }}
        persistence={{ orgId, wizard }}
        onAppChange={onAppChange}
        onBranchChange={onBranchChange}
        onClearError={onClearError}
        onSubmit={() => {}}
      />
    )

    expect(screen.getByText('Install name already in use')).toBeTruthy()
    expect(screen.getByLabelText('Install group label env key')).toHaveProperty(
      'disabled',
      true
    )
    expect(screen.getByLabelText('Label 1 key')).toHaveProperty('value', 'team')
    expect(onAppChange).toHaveBeenCalledWith('app_payments')
    expect(onBranchChange).toHaveBeenCalledWith('branch_main')

    fireEvent.click(screen.getByRole('button', { name: 'Edit install name' }))

    const nameInput = screen.getByLabelText('Install name')
    expect(nameInput).toHaveProperty('value', 'production')
    expect(nameInput).toHaveProperty('disabled', false)
    expect(nameInput.getAttribute('aria-invalid')).toBe('true')
    expect(document.activeElement).toBe(nameInput)
    expect(
      screen.getByText('An install with this name already exists')
    ).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Continue' })).toHaveAttribute(
      'aria-disabled',
      'true'
    )
    expect(onClearError).toHaveBeenCalledTimes(1)

    fireEvent.change(nameInput, { target: { value: 'production-2' } })

    expect(nameInput.getAttribute('aria-invalid')).toBeNull()
    expect(
      screen.queryByText('An install with this name already exists')
    ).toBeNull()
    expect(
      loadDraft<{
        values: { labels: Array<{ key: string; value: string }> }
      }>(orgId, wizard)?.values.values.labels
    ).toEqual([{ key: 'team', value: 'payments' }])
  })
})
