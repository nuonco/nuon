import { describe, expect, test } from 'vitest'
import {
  toSentenceCase,
  toTitleCase,
  getInitials,
  kebabToWords,
  snakeToWords,
  slugify,
  getParentPath,
  formatBytes,
  getFlagEmoji,
} from './string-utils'

describe('string-utils', () => {
  describe('toSentenceCase', () => {
    test('should capitalize first letter', () => {
      expect(toSentenceCase('hello world')).toBe('Hello world')
    })

    test('should handle empty string', () => {
      expect(toSentenceCase('')).toBe('')
      expect(toSentenceCase()).toBe('')
    })

    test('should lowercase remaining letters', () => {
      expect(toSentenceCase('HELLO WORLD')).toBe('Hello world')
    })
  })

  describe('toTitleCase', () => {
    test('should convert to title case', () => {
      expect(toTitleCase('hello world')).toBe('Hello World')
    })

    test('should handle dashes and underscores', () => {
      expect(toTitleCase('hello-world_foo')).toBe('Hello World Foo')
    })

    test('should handle empty string', () => {
      expect(toTitleCase('')).toBe('')
    })
  })

  describe('getInitials', () => {
    test('should get initials from full name', () => {
      expect(getInitials('John Doe')).toBe('JD')
    })

    test('should handle single word', () => {
      expect(getInitials('Alice')).toBe('A')
    })

    test('should handle underscores and dashes', () => {
      expect(getInitials('jane_doe')).toBe('JD')
      expect(getInitials('bob-smith')).toBe('BS')
    })

    test('should handle empty string', () => {
      expect(getInitials('')).toBe('')
      expect(getInitials()).toBe('')
    })
  })

  describe('kebabToWords', () => {
    test('should convert kebab-case to words', () => {
      expect(kebabToWords('foo-bar-baz')).toBe('foo bar baz')
    })

    test('should handle empty string', () => {
      expect(kebabToWords('')).toBe('')
    })
  })

  describe('snakeToWords', () => {
    test('should convert snake_case to words', () => {
      expect(snakeToWords('foo_bar_baz')).toBe('foo bar baz')
    })

    test('should handle empty string', () => {
      expect(snakeToWords('')).toBe('')
    })
  })

  describe('slugify', () => {
    test('should create URL-safe slug', () => {
      expect(slugify('Hello World!')).toBe('hello-world')
    })

    test('should handle multiple spaces', () => {
      expect(slugify('foo   bar')).toBe('foo-bar')
    })

    test('should remove special characters', () => {
      expect(slugify('Hello@World#Test')).toBe('helloworldtest')
    })

    test('should handle empty string', () => {
      expect(slugify('')).toBe('')
    })
  })

  describe('getParentPath', () => {
    test('should get parent path', () => {
      expect(getParentPath('/foo/bar/baz')).toBe('/foo/bar')
    })

    test('should handle trailing slash', () => {
      expect(getParentPath('/foo/bar/')).toBe('/foo')
    })

    test('should return root for top-level path', () => {
      expect(getParentPath('/foo')).toBe('/')
    })

    test('should handle root path', () => {
      expect(getParentPath('/')).toBe('/')
    })
  })

  describe('formatBytes', () => {
    test('should format bytes', () => {
      expect(formatBytes(500)).toBe('500 Bytes')
    })

    test('should format KB', () => {
      expect(formatBytes(1024)).toBe('1.00 KB')
    })

    test('should format MB', () => {
      expect(formatBytes(1048576)).toBe('1.00 MB')
    })

    test('should format GB', () => {
      expect(formatBytes(1073741824)).toBe('1.00 GB')
    })
  })

  describe('getFlagEmoji', () => {
    test('should return US flag for "us"', () => {
      expect(getFlagEmoji('us')).toBe('🇺🇸')
    })

    test('should handle lowercase', () => {
      expect(getFlagEmoji('ca')).toBe('🇨🇦')
    })

    test('should default to US flag', () => {
      expect(getFlagEmoji()).toBe('🇺🇸')
    })
  })
})
