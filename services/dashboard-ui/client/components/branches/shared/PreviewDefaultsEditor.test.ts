import { describe, expect, test } from 'bun:test'
import {
  defaultPreviewDefaults,
  previewDefaultsFromConfig,
  previewDefaultsToConfig,
} from './PreviewDefaultsEditor'

describe('preview defaults', () => {
  test('uses none when preview config is omitted', () => {
    expect(defaultPreviewDefaults().mode).toBe('none')
    expect(previewDefaultsFromConfig(undefined).mode).toBe('none')
  })

  test('serializes none without an install target', () => {
    expect(
      previewDefaultsToConfig(
        {
          ...defaultPreviewDefaults(),
          installId: 'install-1',
        },
        []
      )
    ).toEqual({ mode: 'none' })
  })
})
