import { describe, expect, test } from 'bun:test'
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
  toOrdinal,
  indexToOrdinal,
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

  describe('toOrdinal', () => {
    test('should handle 1st, 2nd, 3rd', () => {
      expect(toOrdinal(1)).toBe('1st')
      expect(toOrdinal(2)).toBe('2nd')
      expect(toOrdinal(3)).toBe('3rd')
    })

    test('should handle 4th through 10th', () => {
      expect(toOrdinal(4)).toBe('4th')
      expect(toOrdinal(5)).toBe('5th')
      expect(toOrdinal(10)).toBe('10th')
    })

    test('should handle teen exceptions (11th, 12th, 13th)', () => {
      expect(toOrdinal(11)).toBe('11th')
      expect(toOrdinal(12)).toBe('12th')
      expect(toOrdinal(13)).toBe('13th')
    })

    test('should handle 21st, 22nd, 23rd', () => {
      expect(toOrdinal(21)).toBe('21st')
      expect(toOrdinal(22)).toBe('22nd')
      expect(toOrdinal(23)).toBe('23rd')
    })

    test('should handle larger numbers', () => {
      expect(toOrdinal(101)).toBe('101st')
      expect(toOrdinal(112)).toBe('112th')
      expect(toOrdinal(1000)).toBe('1000th')
    })
  })

  describe('indexToOrdinal', () => {
    test('should convert index 0 to 1st', () => {
      expect(indexToOrdinal(0)).toBe('1st')
    })

    test('should convert index 1 to 2nd', () => {
      expect(indexToOrdinal(1)).toBe('2nd')
    })

    test('should convert index 2 to 3rd', () => {
      expect(indexToOrdinal(2)).toBe('3rd')
    })

    test('should handle larger indices', () => {
      expect(indexToOrdinal(10)).toBe('11th')
      expect(indexToOrdinal(20)).toBe('21st')
      expect(indexToOrdinal(99)).toBe('100th')
    })
  })
})
