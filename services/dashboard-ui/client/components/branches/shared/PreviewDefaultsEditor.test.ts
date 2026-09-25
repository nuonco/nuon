import { describe, expect, test } from 'bun:test'
import {
  defaultPreviewDefaults,
  previewDefaultsFromConfig,
  previewDefaultsToConfig,
} from './PreviewDefaultsEditor'

describe('preview defaults', () => {
  test('uses plan-only when preview config is omitted', () => {
    expect(defaultPreviewDefaults().mode).toBe('plan-only')
    expect(previewDefaultsFromConfig(undefined).mode).toBe('plan-only')
  })

  test('serializes none without an install target', () => {
    expect(
      previewDefaultsToConfig(
        {
          ...defaultPreviewDefaults(),
          mode: 'none',
          installId: 'install-1',
        },
        []
      )
    ).toEqual({ mode: 'none' })
  })
})
