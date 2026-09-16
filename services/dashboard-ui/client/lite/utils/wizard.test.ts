import { describe, expect, test } from 'bun:test'
import {
  isWizardComplete,
  resolveWizardStep,
  type IWizardDescriptor,
} from './wizard'

interface IState {
  name?: string
  connected?: boolean
  provisioned?: boolean
}

const descriptor: IWizardDescriptor<IState> = {
  steps: [
    {
      id: 'name',
      label: 'Name',
      complete: (state) => Boolean(state.name),
      render: () => null,
    },
    {
      id: 'connect',
      label: 'Connect',
      complete: (state) => Boolean(state.connected),
      render: () => null,
    },
    {
      id: 'provision',
      label: 'Provision',
      complete: (state) => Boolean(state.provisioned),
      render: () => null,
    },
  ],
}

describe('resolveWizardStep', () => {
  test('selects the first incomplete step as current', () => {
    const resolved = resolveWizardStep(descriptor, { name: 'Payments' })

    expect(resolved.finished).toBe(false)
    expect(resolved.current?.id).toBe('connect')
    expect(resolved.currentIndex).toBe(1)
    expect(resolved.statuses).toEqual(['done', 'current', 'upcoming'])
  })

  test('starts at the first step when nothing is complete', () => {
    const resolved = resolveWizardStep(descriptor, {})

    expect(resolved.current?.id).toBe('name')
    expect(resolved.statuses).toEqual(['current', 'upcoming', 'upcoming'])
  })

  test('reports finished when every step is complete', () => {
    const resolved = resolveWizardStep(descriptor, {
      name: 'Payments',
      connected: true,
      provisioned: true,
    })

    expect(resolved.finished).toBe(true)
    expect(resolved.current).toBeUndefined()
    expect(resolved.statuses).toEqual(['done', 'done', 'done'])
  })

  test('moves the current step back when earlier state regresses', () => {
    const resolved = resolveWizardStep(descriptor, {
      connected: true,
      provisioned: true,
    })

    expect(resolved.current?.id).toBe('name')
    expect(resolved.statuses).toEqual(['current', 'upcoming', 'upcoming'])
  })
})

describe('isWizardComplete', () => {
  test('is true only when every step is complete', () => {
    expect(isWizardComplete(descriptor, { name: 'Payments' })).toBe(false)
    expect(
      isWizardComplete(descriptor, {
        name: 'Payments',
        connected: true,
        provisioned: true,
      })
    ).toBe(true)
  })
})
