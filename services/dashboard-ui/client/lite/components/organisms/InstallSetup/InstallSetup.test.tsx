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
    const onDiscard = mock()

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
        onDiscard={onDiscard}
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

    fireEvent.click(screen.getByRole('button', { name: '2. Install details' }))
    expect(screen.getByLabelText('Install name')).toHaveProperty(
      'disabled',
      true
    )

    fireEvent.click(screen.getByRole('button', { name: 'Edit step' }))
    expect(screen.getByLabelText('Install name')).toHaveProperty(
      'disabled',
      false
    )
    expect(screen.queryByRole('button', { name: 'Continue' })).not.toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Done editing' }))
    expect(screen.getByLabelText('Install name')).toHaveProperty(
      'disabled',
      true
    )

    fireEvent.click(
      screen.getByRole('button', { name: '3. Labels and inputs' })
    )
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

    fireEvent.click(screen.getByRole('button', { name: 'Discard setup' }))
    expect(onDiscard).toHaveBeenCalledTimes(1)
    expect(loadDraft(orgId, wizard)).toBeUndefined()
  })

  test('when app sync is disabled, no branches shows the app-branch empty state and blocks submission without a branch', () => {
    saveDraft(orgId, `${wizard}-flag`, {
      progress: { app: true, details: true },
      values: {
        appId: 'app_payments',
        name: 'production',
        location: 'us-west-2',
        accountId: '',
        autoApprove: false,
        stackOnly: false,
        labels: [],
        inputs: [],
        branchId: '',
        installGroupId: '',
      },
    })

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
        branches={[]}
        inputs={[]}
        platform="aws"
        appSyncDisabled
        configurationReady
        persistence={{ orgId, wizard: `${wizard}-flag` }}
        onAppChange={() => {}}
        onBranchChange={() => {}}
        onSubmit={() => {}}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: '1. App' }))
    expect(
      screen.getByText(
        'This app has no app branches. Create an app branch, then come back to create installs once it has a successful run.'
      )
    ).toBeTruthy()

    fireEvent.click(
      screen.getByRole('button', { name: '3. Labels and inputs' })
    )

    expect(
      screen.getByRole('button', { name: 'Create install' })
    ).toHaveAttribute('aria-disabled', 'true')
  })

  test('flag off shows the branch enrollment caption for optional enrollment', () => {
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
        branches={[]}
        inputs={[]}
        platform="aws"
        configurationReady
        persistence={{ orgId, wizard: `${wizard}-caption-off` }}
        onAppChange={() => {}}
        onBranchChange={() => {}}
        onSubmit={() => {}}
      />
    )

    fireEvent.click(screen.getByText('Choose an app'))
    fireEvent.click(screen.getByRole('option', { name: /Payments API/ }))

    expect(
      screen.getByText(
        'Join an install group now, or enroll this install later.'
      )
    ).toBeTruthy()
    expect(
      screen.queryByText('Choose an app branch for this install.')
    ).toBeNull()
  })

  test('when app sync is disabled, shows the branch enrollment caption for required enrollment', () => {
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
        branches={[]}
        inputs={[]}
        platform="aws"
        appSyncDisabled
        configurationReady
        persistence={{ orgId, wizard: `${wizard}-caption-on` }}
        onAppChange={() => {}}
        onBranchChange={() => {}}
        onSubmit={() => {}}
      />
    )

    fireEvent.click(screen.getByText('Choose an app'))
    fireEvent.click(screen.getByRole('option', { name: /Payments API/ }))

    expect(
      screen.getByText('Choose an app branch for this install.')
    ).toBeTruthy()
    expect(
      screen.queryByText(
        'Join an install group now, or enroll this install later.'
      )
    ).toBeNull()
  })
})
