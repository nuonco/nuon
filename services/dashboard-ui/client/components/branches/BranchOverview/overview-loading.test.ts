import { describe, expect, test } from 'bun:test'
import type { TCompositeError, TInstallWorkflowStep } from '@/types'
import {
  buildOverviewLoadingStages,
  fetchCommitReady,
  installFailureHref,
  overviewCompositeError,
  preRolloutCompositeError,
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
      ['starting', 'in-progress'],
      ['fetch-commit', 'pending'],
      ['app-config', 'pending'],
      ['build-components', 'pending'],
      ['rollout', 'pending'],
    ])
  })

  test('drops the rollout stage for a build and validate preview', () => {
    const stages = buildOverviewLoadingStages({
      steps: [step('building components and sandbox', 'success')],
      sha: 'abc',
      previewMode: 'build-only',
    })
    expect(stages.map((stage) => stage.id)).toEqual([
      'starting',
      'fetch-commit',
      'app-config',
      'build-components',
    ])
  })

  test('ends a plan preview on the plan stage', () => {
    const stages = buildOverviewLoadingStages({
      steps: [step('plan preview install', 'in-progress')],
      previewMode: 'plan-only',
    })
    expect(stages.at(-1)).toMatchObject({ id: 'plan', status: 'in-progress' })
  })

  test('ends an apply preview on the apply stage', () => {
    const stages = buildOverviewLoadingStages({
      steps: [step('apply preview install', 'success')],
      previewMode: 'apply',
    })
    expect(stages.at(-1)).toMatchObject({ id: 'apply', status: 'success' })
  })

  test('summarizes install groups in the rollout stage', () => {
    const rollout = (groupStatuses: string[]) =>
      buildOverviewLoadingStages({ steps: [], groupStatuses }).find(
        (stage) => stage.id === 'rollout'
      )?.status

    expect(rollout(['pending', 'pending'])).toBe('pending')
    expect(rollout(['success', 'approval-awaiting'])).toBe('in-progress')
    expect(rollout(['success', 'pending'])).toBe('in-progress')
    expect(rollout(['success', 'auto-skipped'])).toBe('success')
    expect(rollout(['success', 'error', 'pending'])).toBe('error')
  })

  test('shows the commit only after fetch commit succeeds and a sha exists', () => {
    const fetching = buildOverviewLoadingStages({
      steps: [step('fetch commit', 'in-progress')],
      sha: 'abc',
    })
    expect(fetching.find((stage) => stage.id === 'fetch-commit')?.status).toBe(
      'in-progress'
    )
    expect(
      buildOverviewLoadingStages({
        steps: [step('fetch commit', 'success')],
      }).find((stage) => stage.id === 'fetch-commit')?.status
    ).toBe('in-progress')
    expect(fetchCommitReady([step('fetch commit', 'in-progress')], 'abc')).toBe(
      false
    )

    const ready = buildOverviewLoadingStages({
      steps: [step('fetch commit', 'success')],
      sha: 'abc',
    })
    expect(ready.find((stage) => stage.id === 'starting')?.status).toBe(
      'success'
    )
    expect(ready.find((stage) => stage.id === 'fetch-commit')?.status).toBe(
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

  test('uses a failed install group deploy when earlier steps succeeded', () => {
    const error = overviewCompositeError([
      step('fetch commit', 'success'),
      step('fetch app config', 'success'),
      step('build components', 'success'),
      step('deploy install group: canary', 'error', {
        message: 'jm-test-001 failed during deploy',
      }),
    ])
    expect(error?.message).toBe('jm-test-001 failed during deploy')
  })

  test('prefers the app branch run composite error', () => {
    const runError = {
      message: 'delta failed during deploy',
    } as TCompositeError
    const error = overviewCompositeError(
      [step('fetch app config', 'error', { message: 'Config failed' })],
      runError
    )
    expect(error?.message).toBe('delta failed during deploy')
  })

  test('keeps pre-rollout failures for the page banner', () => {
    const error = preRolloutCompositeError([
      step('fetch commit', 'success'),
      step('build components', 'error', { message: 'Build failed' }),
    ])
    expect(error?.message).toBe('Build failed')
  })

  test('leaves rollout install failures to the install group cards', () => {
    const steps = [
      step('fetch commit', 'success'),
      step('fetch app config', 'success'),
      step('build components', 'success'),
      step('deploy install group: canary', 'error', {
        message: 'acme-prod failed during deploy',
      }),
    ]
    expect(preRolloutCompositeError(steps)).toBeUndefined()
    expect(
      preRolloutCompositeError(steps.slice(0, 3), {
        type: 'install_group.install_update_failed',
        message: 'acme-prod failed during deploy',
      } as TCompositeError)
    ).toBeUndefined()
  })

  test('links an install update failure to the install workflow', () => {
    const error = {
      type: 'install_group.install_update_failed',
      message: 'jm-test-001 failed during deploy',
      data: { install_id: 'ins_1', workflow_id: 'wf_1' },
    } as unknown as TCompositeError
    expect(installFailureHref(error, 'org_acme')).toBe(
      '/org_acme/installs/ins_1/workflows/wf_1'
    )
    expect(installFailureHref({ message: 'other' }, 'org_acme')).toBeUndefined()
  })
})
