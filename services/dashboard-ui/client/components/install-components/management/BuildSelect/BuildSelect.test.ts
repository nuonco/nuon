import { describe, expect, test } from 'bun:test'
import type { TBuild } from '@/types'
import { excludePreviewBuilds } from './BuildSelect'

describe('excludePreviewBuilds', () => {
  test('removes preview builds from deploy choices', () => {
    const builds = [
      { id: 'bld-preview', is_preview: true },
      { id: 'bld-main', is_preview: false },
      { id: 'bld-legacy' },
    ] as TBuild[]

    expect(excludePreviewBuilds(builds).map((build) => build.id)).toEqual([
      'bld-main',
      'bld-legacy',
    ])
  })
})
