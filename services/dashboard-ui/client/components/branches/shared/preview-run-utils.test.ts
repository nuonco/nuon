import { describe, expect, test } from 'bun:test'
import type { TAppBranchRun } from '@/types'
import { previewSourceLabel } from './preview-run-utils'

describe('previewSourceLabel', () => {
  test('distinguishes a preview git ref from the Nuon branch name', () => {
    const run = {
      app_branch: { name: 'production' },
      preview: { source: 'branch', git_ref: 'feature/payments' },
    } as TAppBranchRun

    expect(previewSourceLabel(run)).toBe('git ref: feature/payments')
  })

  test('does not repeat the Nuon branch name as a source badge', () => {
    const run = {
      app_branch: { name: 'production' },
      preview: { source: 'branch', git_ref: 'production' },
    } as TAppBranchRun

    expect(previewSourceLabel(run)).toBeUndefined()
  })
})
