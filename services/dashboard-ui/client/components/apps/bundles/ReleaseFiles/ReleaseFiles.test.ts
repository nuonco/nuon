import { describe, expect, test } from 'bun:test'
import {
  maxReleaseFilePreviewBytes,
  releaseFileEntryCanPreview,
  type TReleaseFileEntry,
} from './ReleaseFiles'

const entryWithSizes = (
  currentSize: number,
  previousSize?: number
): TReleaseFileEntry => ({
  category: 'source',
  change: previousSize === undefined ? 'added' : 'modified',
  current: { size: currentSize },
  path: 'large.rego',
  previous: previousSize === undefined ? undefined : { size: previousSize },
})

describe('releaseFileEntryCanPreview', () => {
  test('allows files at the preview limit', () => {
    expect(
      releaseFileEntryCanPreview(entryWithSizes(maxReleaseFilePreviewBytes))
    ).toBe(true)
  })

  test('rejects either side of a diff above the preview limit', () => {
    expect(
      releaseFileEntryCanPreview(
        entryWithSizes(
          maxReleaseFilePreviewBytes,
          maxReleaseFilePreviewBytes + 1
        )
      )
    ).toBe(false)
  })
})
