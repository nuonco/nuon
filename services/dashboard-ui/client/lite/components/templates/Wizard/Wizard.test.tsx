import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import type { IWizardDescriptor } from '../../../utils/wizard'
import { Wizard } from './Wizard'

afterEach(cleanup)

interface IState {
  name: string
  connected: boolean
  provisioned: boolean
}

const descriptor = (
  onNameChange: (value: string) => void = () => {}
): IWizardDescriptor<IState> => ({
  steps: [
    {
      id: 'name',
      label: 'Name',
      complete: (state) => state.name.trim().length > 0,
      render: ({ state, readOnly }) => (
        <input
          aria-label="Name"
          value={state.name}
          disabled={readOnly}
          onChange={(event) => onNameChange(event.target.value)}
        />
      ),
    },
    {
      id: 'connect',
      label: 'Connect',
      complete: (state) => state.connected,
      render: ({ readOnly }) => (
        <p>{readOnly ? 'Connected' : 'Connect the branch'}</p>
      ),
    },
    {
      id: 'provision',
      label: 'Provision',
      complete: (state) => state.provisioned,
      render: ({ state }) => (
        <p>{state.provisioned ? 'Provision finished' : 'Waiting'}</p>
      ),
    },
  ],
})

const renderWizard = (state: IState) =>
  render(
    <Wizard
      descriptor={descriptor()}
      state={state}
      exitAction={{ children: 'Continue to app' }}
    />
  )

describe('Wizard', () => {
  test('opens the first incomplete step and lists the rest', () => {
    renderWizard({ name: 'Payments', connected: false, provisioned: false })

    expect(screen.getByText('Connect the branch')).toBeTruthy()
    expect(screen.queryByLabelText('Name')).toBeNull()
    expect(screen.getByText('Step 2 of 3')).toBeTruthy()
    expect(
      screen.getByRole('button', { name: 'Continue to app' })
    ).toHaveAttribute('aria-disabled', 'true')
    expect(screen.getByRole('button', { name: '3. Provision' })).toHaveProperty(
      'disabled',
      true
    )
  })

  test('follows the current step forward as state progresses', () => {
    const view = renderWizard({
      name: 'Payments',
      connected: false,
      provisioned: false,
    })

    expect(screen.getByText('Connect the branch')).toBeTruthy()

    view.rerender(
      <Wizard
        descriptor={descriptor()}
        state={{ name: 'Payments', connected: true, provisioned: false }}
        exitAction={{ children: 'Continue to app' }}
      />
    )

    expect(screen.getByText('Waiting')).toBeTruthy()
    expect(screen.getByText('Step 3 of 3')).toBeTruthy()
  })

  test('returns to the current step after revisiting a done one', () => {
    renderWizard({ name: 'Payments', connected: true, provisioned: false })

    fireEvent.click(screen.getByRole('button', { name: '1. Name' }))
    expect(screen.getByLabelText('Name')).toHaveProperty('disabled', true)

    fireEvent.click(screen.getByRole('button', { name: '3. Provision' }))
    expect(screen.getByText('Waiting')).toBeTruthy()
  })

  test('lets a finished watcher wait on the exit action', () => {
    renderWizard({ name: 'Payments', connected: true, provisioned: true })

    expect(screen.getByText('Provision finished')).toBeTruthy()
    expect(
      screen.getByRole('button', { name: 'Continue to app' })
    ).not.toHaveAttribute('aria-disabled')
  })

  test('moves back when earlier state regresses', () => {
    renderWizard({ name: '', connected: true, provisioned: true })

    expect(screen.getByLabelText('Name')).toHaveProperty('disabled', false)
    expect(screen.getByText('Step 1 of 3')).toBeTruthy()
  })

  test('edits a completed step and returns it to read-only', () => {
    const onNameChange = () => {}
    render(
      <Wizard
        descriptor={descriptor(onNameChange)}
        state={{ name: 'Payments', connected: true, provisioned: false }}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: '1. Name' }))
    expect(screen.getByLabelText('Name')).toHaveProperty('disabled', true)

    fireEvent.click(screen.getByRole('button', { name: 'Edit step' }))
    expect(screen.getByLabelText('Name')).toHaveProperty('disabled', false)

    fireEvent.click(screen.getByRole('button', { name: 'Done editing' }))
    expect(screen.getByLabelText('Name')).toHaveProperty('disabled', true)
  })

  test('runs the discard action', () => {
    let discarded = false
    render(
      <Wizard
        descriptor={descriptor()}
        state={{ name: '', connected: false, provisioned: false }}
        discardAction={{
          children: 'Discard setup',
          onClick: () => {
            discarded = true
          },
        }}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: 'Discard setup' }))
    expect(discarded).toBe(true)
  })
})
