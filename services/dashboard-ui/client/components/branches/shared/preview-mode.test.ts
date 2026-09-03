import { describe, expect, test } from 'bun:test'
import { previewModeDisplayLabel } from './preview-mode'

describe('previewModeDisplayLabel', () => {
  test.each([
    ['build-only', 'Build and validate'],
    ['plan-only', 'Plan only'],
    ['apply', 'Apply'],
  ] as const)('labels %s as %s', (mode, label) => {
    expect(previewModeDisplayLabel(mode)).toBe(label)
  })
})
