import { describe, expect, test } from 'bun:test'
import type { TWorkflow } from '@/types'
import { branchRunOutcome } from './run-outcome'

const workflow = (
  steps: NonNullable<TWorkflow['steps']>
): TWorkflow => ({ steps }) as TWorkflow

const deploy = (
  name: string,
  stepStatus: string,
  installs: string[]
) =>
  ({
    name: `deploy install group: ${name}`,
    status: {
      status: stepStatus,
      metadata: {
        installs: installs.map((status) => ({ status })),
      },
    },
  }) as NonNullable<TWorkflow['steps']>[number]

describe('branchRunOutcome', () => {
  test('counts finished installs', () => {
    expect(
      branchRunOutcome(
        workflow([
          deploy('Canary', 'success', ['success', 'success']),
          deploy('Primary region', 'success', ['success']),
        ])
      )
    ).toEqual({ text: '3 installs updated', failed: false })
  })

  test('names the group that failed', () => {
    expect(
      branchRunOutcome(
        workflow([
          deploy('Canary', 'success', ['success']),
          deploy('Primary region', 'error', ['success', 'error']),
        ])
      )
    ).toEqual({ text: 'Failed in Primary region', failed: true })
  })

  test('shows progress while a group is still deploying', () => {
    expect(
      branchRunOutcome(
        workflow([
          deploy('Canary', 'success', ['success', 'success']),
          deploy('Primary region', 'in-progress', ['success', 'in-progress']),
        ])
      )
    ).toEqual({ text: '3 of 4 installs updated', failed: false })
  })

  test('says nothing before a deploy group starts', () => {
    expect(
      branchRunOutcome(
        workflow([
          {
            name: 'building components and sandbox',
            status: { status: 'in-progress' },
          } as NonNullable<TWorkflow['steps']>[number],
        ])
      )
    ).toBeUndefined()
  })
})
