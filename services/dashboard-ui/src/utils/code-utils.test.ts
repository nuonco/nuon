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

    test('should handle JSON stringify errors gracefully', () => {
      const circular: any = {}
      circular.self = circular

      const result = diffLines(circular, 'test')
      expect(typeof result).toBe('string')
    })
  })
})
