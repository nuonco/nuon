import { describe, expect, test } from 'bun:test'
import type { TWorkflow } from '@/types'
import { previousBranchRun } from './InstallRolloutPanel'

describe('previousBranchRun', () => {
  test('skips the run being rolled out and workflows without a branch run', () => {
    const workflows = [
      { id: 'wf-3', app_branch_runs: [{ id: 'run-current', head_sha: 'ccc' }] },
      { id: 'wf-2' },
      { id: 'wf-1', app_branch_runs: [{ id: 'run-old', head_sha: 'aaa' }] },
    ] as TWorkflow[]
    expect(previousBranchRun(workflows, 'run-current')?.id).toBe('run-old')
    expect(previousBranchRun([], 'run-current')).toBeUndefined()
  })
})
