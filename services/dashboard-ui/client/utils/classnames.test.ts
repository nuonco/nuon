import { describe, expect, test } from 'vitest'
import { cn } from './classnames'

describe('classnames', () => {
  describe('cn', () => {
    test('should combine string class names', () => {
      expect(cn('foo', 'bar')).toBe('foo bar')
    })

    test('should handle undefined values', () => {
      expect(cn('foo', undefined, 'bar')).toBe('foo bar')
    })

    test('should handle conditional object syntax', () => {
      expect(cn('base', { active: true, disabled: false })).toBe('base active')
    })

    test('should handle mixed inputs', () => {
      expect(cn('base', { active: true }, undefined, 'extra')).toBe(
        'base active extra'
      )
    })

    test('should handle empty input', () => {
      expect(cn()).toBe('')
    })

    test('should handle all undefined inputs', () => {
      expect(cn(undefined, undefined)).toBe('')
    })

    test('should handle complex conditional logic', () => {
      const isActive = true
      const isDisabled = false
      const variant = 'primary'

      expect(
        cn('button', variant && `button--${variant}`, {
          'button--active': isActive,
          'button--disabled': isDisabled,
        })
      ).toBe('button button--primary button--active')
    })
  })
})
