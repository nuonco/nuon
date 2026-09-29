import { describe, expect, test } from 'bun:test'
import type { TInstallWorkflowStep } from '@/types'
import {
  buildOverviewLoadingStages,
  fetchCommitReady,
  overviewCompositeError,
} from './overview-loading'

const step = (
  name: string,
  status: string,
  compositeError?: { message: string }
): TInstallWorkflowStep =>
  ({
    id: name,
    name,
    status: { status, composite_error: compositeError },
  }) as TInstallWorkflowStep

describe('buildOverviewLoadingStages', () => {
  test('waits for the workflow when no steps exist', () => {
    expect(
      buildOverviewLoadingStages({ steps: [] }).map((stage) => [
        stage.id,
        stage.status,
      ])
    ).toEqual([
      ['waiting', 'in-progress'],
      ['fetch-commit', 'pending'],
      ['show-commit', 'pending'],
      ['app-config', 'pending'],
      ['build-components', 'pending'],
    ])
  })

  test('shows the commit only after fetch commit succeeds and a sha exists', () => {
    const fetching = buildOverviewLoadingStages({
      steps: [step('fetch commit', 'in-progress')],
      sha: 'abc',
    })
    expect(fetching.find((stage) => stage.id === 'show-commit')?.status).toBe(
      'pending'
    )
    expect(
      fetchCommitReady([step('fetch commit', 'in-progress')], 'abc')
    ).toBe(false)

    const ready = buildOverviewLoadingStages({
      steps: [step('fetch commit', 'success')],
      sha: 'abc',
    })
    expect(ready.find((stage) => stage.id === 'waiting')?.status).toBe(
      'success'
    )
    expect(ready.find((stage) => stage.id === 'show-commit')?.status).toBe(
      'success'
    )
    expect(fetchCommitReady([step('fetch commit', 'success')], 'abc')).toBe(
      true
    )
    expect(fetchCommitReady([step('fetch commit', 'success')])).toBe(false)
  })

  test('keeps build app config separate from component builds', () => {
    const stages = buildOverviewLoadingStages({
      steps: [
        step('fetch commit', 'success'),
        step('sync app config', 'in-progress'),
        step('build components', 'pending'),
      ],
      sha: 'abc',
    })
    expect(stages.find((stage) => stage.id === 'app-config')?.status).toBe(
      'in-progress'
    )
    expect(
      stages.find((stage) => stage.id === 'build-components')?.status
    ).toBe('pending')
  })

  test('returns the first failing early step composite error', () => {
    const error = overviewCompositeError([
      step('fetch commit', 'success'),
      step('fetch app config', 'error', { message: 'Config failed' }),
      step('build components', 'error', { message: 'Build failed' }),
    ])
    expect(error?.message).toBe('Config failed')
  })
})
