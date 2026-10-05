import { expect, test } from 'bun:test'
import type { TBranchRunComparisonConfigDiff } from '@/lib'
import {
  scopedComparisonConfigDiff,
  scopedComparisonFiles,
} from './comparison-scope'

const config: TBranchRunComparisonConfigDiff = {
  additions: 1,
  removals: 1,
  changed: 3,
  sections: [
    {
      name: 'Stack',
      additions: 0,
      removals: 0,
      changed: 1,
      entries: [{ name: 'Stack', op: 'change', file: 'stack.toml' }],
    },
    {
      name: 'Sandbox',
      additions: 0,
      removals: 0,
      changed: 1,
      entries: [{ name: 'Sandbox', op: 'change', file: 'sandbox.toml' }],
    },
    {
      name: 'Components',
      additions: 1,
      removals: 1,
      changed: 1,
      entries: [
        { name: 'api', op: 'change', file: 'components/api.toml' },
        { name: 'worker', op: 'add', file: 'components/worker.toml' },
        { name: 'legacy', op: 'remove', file: 'components/legacy.toml' },
      ],
    },
  ],
}
const files = [
  'stack.toml',
  'sandbox.toml',
  'components/api.toml',
  'components/worker.toml',
  'components/legacy.toml',
  'shared/values.yaml',
].map((path) => ({ path, op: 'modified' }))

test('stack-only comparisons omit unrelated resource sections and defining files', () => {
  const scope = { sections: ['Stack'] }
  expect(
    scopedComparisonConfigDiff(config, scope)?.sections.map(({ name }) => name)
  ).toEqual(['Stack'])
  expect(
    scopedComparisonFiles(files, config, scope)?.map(({ path }) => path)
  ).toEqual(['stack.toml', 'shared/values.yaml'])
})

test('component-only comparisons include only targeted entities and their counts', () => {
  const scope = { sections: ['Components'], components: ['api', 'worker'] }
  expect(scopedComparisonConfigDiff(config, scope)?.sections).toMatchObject([
    {
      name: 'Components',
      additions: 1,
      removals: 0,
      changed: 1,
      entries: [{ name: 'api' }, { name: 'worker' }],
    },
  ])
  expect(
    scopedComparisonFiles(files, config, scope)?.map(({ path }) => path)
  ).toEqual([
    'components/api.toml',
    'components/worker.toml',
    'shared/values.yaml',
  ])
})

test('mixed scope preserves infrastructure and removals without showing untouched components', () => {
  const diff = scopedComparisonConfigDiff(config, {
    sections: ['Stack', 'Sandbox', 'Components'],
    components: ['api', 'legacy'],
  })
  expect(diff?.sections.map(({ name }) => name)).toEqual([
    'Stack',
    'Sandbox',
    'Components',
  ])
  expect(diff?.sections[2].entries.map(({ name }) => name)).toEqual([
    'api',
    'legacy',
  ])
  expect(diff?.sections[2].removals).toBe(1)
})

test('unscoped branch views and missing attribution retain their original data', () => {
  expect(scopedComparisonConfigDiff(config)).toBe(config)
  expect(scopedComparisonFiles(files, config)).toBe(files)
  expect(scopedComparisonFiles(files, undefined, { sections: ['Stack'] })).toBe(
    files
  )
})
