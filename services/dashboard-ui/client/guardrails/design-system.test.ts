import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { ratchetFailures, scan, type TViolationCounts } from './design-system'

const BASELINE_PATH = join(import.meta.dir, 'design-system.baseline.json')

const baseline: TViolationCounts = JSON.parse(readFileSync(BASELINE_PATH, 'utf8'))

describe('dashboard design-system guardrails', () => {
  test.skip('no new or stale violations outside onboarding, Lite, and playground', () => {
    const failures = ratchetFailures(scan(), baseline)

    expect(
      failures,
      'Each rule is documented in client/components/PATTERNS.md "Choosing a control" and ' +
        'DESIGN.md. onboarding/, lite/, and playground/ are out of scope — see AGENTS.md ' +
        '"Design system".',
    ).toEqual([])
  })
})
