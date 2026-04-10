import { describe, expect, test } from 'vitest'
import { diffLines } from './code-utils'

describe('code-utils', () => {
  describe('diffLines', () => {
    test('should show no diff for identical strings', () => {
      const before = 'line1\nline2\nline3'
      const after = 'line1\nline2\nline3'
      expect(diffLines(before, after)).toBe('  line1\n  line2\n  line3')
    })

    test('should show additions and removals', () => {
      const before = 'line1\nline2'
      const after = 'line1\nline3'
      const result = diffLines(before, after)
      expect(result).toContain('  line1')
      expect(result).toContain('- line2')
      expect(result).toContain('+ line3')
    })

    test('should handle empty strings', () => {
      expect(diffLines('', '')).toBe('No diff to show')
    })

    test('should handle null and undefined inputs', () => {
      expect(diffLines(null, null)).toBe('No diff to show')
      expect(diffLines(undefined, undefined)).toBe('No diff to show')
      expect(diffLines(null, 'test')).toBe('+ test')
      expect(diffLines('test', null)).toBe('- test')
    })

    test('should handle objects by JSON stringifying them', () => {
      const before = { name: 'John', age: 30 }
      const after = { name: 'John', age: 31 }
      const result = diffLines(before, after)

      expect(result).toContain('  {')
      expect(result).toContain('    "name": "John",')
      expect(result).toContain('- ')
      expect(result).toContain('+ ')
    })

    test('should handle arrays', () => {
      const before = [1, 2, 3]
      const after = [1, 2, 4]
      const result = diffLines(before, after)

      expect(result).toContain('  [')
      expect(result).toContain('    1,')
      expect(result).toContain('    2,')
    })

    test('should handle mixed types', () => {
      const before = 'simple string'
      const after = { complex: 'object' }
      const result = diffLines(before, after)

      expect(result).toContain('- simple string')
      expect(result).toContain('+ {')
    })

    test('should handle different line lengths', () => {
      const before = 'line1'
      const after = 'line1\nline2\nline3'
      const result = diffLines(before, after)

      expect(result).toContain('  line1')
      expect(result).toContain('+ line2')
      expect(result).toContain('+ line3')
    })

    test('should use ~ for YAML key-value changes with same key', () => {
      const before = 'name: foo\nimage: nginx:1.0\nport: 80'
      const after = 'name: foo\nimage: nginx:2.0\nport: 80'
      const result = diffLines(before, after)
      expect(result).toContain('  name: foo')
      expect(result).toContain('~ image: nginx:1.0 -> nginx:2.0')
      expect(result).toContain('  port: 80')
      expect(result).not.toContain('- image')
      expect(result).not.toContain('+ image')
    })

    test('should use ~ for indented YAML key changes', () => {
      const before = 'spec:\n  replicas: 1\n  memory: 256Mi'
      const after = 'spec:\n  replicas: 3\n  memory: 512Mi'
      const result = diffLines(before, after)
      expect(result).toContain('  spec:')
      expect(result).toContain('~   replicas: 1 -> 3')
      expect(result).toContain('~   memory: 256Mi -> 512Mi')
    })

    test('should fall back to +/- when keys differ', () => {
      const before = 'name: foo'
      const after = 'image: bar'
      const result = diffLines(before, after)
      expect(result).toContain('- name: foo')
      expect(result).toContain('+ image: bar')
    })

    test('should not offset lines after a single removal', () => {
      const before = 'line1\nline2\nremoved-line\nline3\nline4'
      const after = 'line1\nline2\nline3\nline4'
      const result = diffLines(before, after)
      expect(result).toContain('  line1')
      expect(result).toContain('  line2')
      expect(result).toContain('- removed-line')
      expect(result).toContain('  line3')
      expect(result).toContain('  line4')
      expect(result).not.toContain('+ line3')
      expect(result).not.toContain('+ line4')
    })

    test('should not offset lines after a single addition', () => {
      const before = 'line1\nline2\nline3'
      const after = 'line1\nline2\nnew-line\nline3'
      const result = diffLines(before, after)
      expect(result).toContain('  line1')
      expect(result).toContain('  line2')
      expect(result).toContain('+ new-line')
      expect(result).toContain('  line3')
      expect(result).not.toContain('- line3')
    })

    test('should handle JSON stringify errors gracefully', () => {
      const circular: any = {}
      circular.self = circular

      const result = diffLines(circular, 'test')
      expect(typeof result).toBe('string')
    })
  })
})
