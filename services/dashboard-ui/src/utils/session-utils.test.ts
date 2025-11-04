import { describe, expect, test } from 'vitest'
import type { User } from '@auth0/nextjs-auth0/types'
import { isNuonSession } from './session-utils'

describe('session-utils', () => {
  describe('isNuonSession', () => {
    test('should return true for valid Nuon email', () => {
      const user: User = {
        email: 'john.doe@nuon.co',
        name: 'John Doe',
        sub: 'auth0|123',
      }
      expect(isNuonSession(user)).toBe(true)
    })

    test('should return true for Nuon email with different format', () => {
      const user: User = {
        email: 'jane.smith@nuon.co',
        name: 'Jane Smith',
        sub: 'auth0|456',
      }
      expect(isNuonSession(user)).toBe(true)
    })

    test('should return false for non-Nuon email', () => {
      const user: User = {
        email: 'user@example.com',
        name: 'External User',
        sub: 'auth0|789',
      }
      expect(isNuonSession(user)).toBe(false)
    })

    test('should return false for different company domain', () => {
      const user: User = {
        email: 'user@company.com',
        name: 'Company User',
        sub: 'auth0|012',
      }
      expect(isNuonSession(user)).toBe(false)
    })

    test('should return false for similar but incorrect domain', () => {
      const user: User = {
        email: 'user@nuon.com',
        name: 'User',
        sub: 'auth0|345',
      }
      expect(isNuonSession(user)).toBe(false)
    })

    test('should return false for domain containing nuon.co', () => {
      const user: User = {
        email: 'user@mynuon.co.com',
        name: 'User',
        sub: 'auth0|678',
      }
      expect(isNuonSession(user)).toBe(false)
    })

    test('should handle user with undefined email', () => {
      const user: User = {
        email: undefined,
        name: 'User',
        sub: 'auth0|901',
      }
      expect(isNuonSession(user)).toBe(false)
    })

    test('should handle user with null email', () => {
      const user: User = {
        email: null as any,
        name: 'User',
        sub: 'auth0|234',
      }
      expect(isNuonSession(user)).toBe(false)
    })

    test('should handle user with empty email', () => {
      const user: User = {
        email: '',
        name: 'User',
        sub: 'auth0|567',
      }
      expect(isNuonSession(user)).toBe(false)
    })

    test('should handle malformed email without @ symbol', () => {
      const user: User = {
        email: 'usernuon.co',
        name: 'User',
        sub: 'auth0|890',
      }
      expect(isNuonSession(user)).toBe(false)
    })

    test('should handle email with multiple @ symbols', () => {
      const user: User = {
        email: 'user@test@nuon.co',
        name: 'User',
        sub: 'auth0|123',
      }
      expect(isNuonSession(user)).toBe(true)
    })

    test('should be case sensitive for domain', () => {
      const user: User = {
        email: 'user@NUON.CO',
        name: 'User',
        sub: 'auth0|456',
      }
      expect(isNuonSession(user)).toBe(false)
    })

    test('should handle email with whitespace', () => {
      const user: User = {
        email: ' user@nuon.co ',
        name: 'User',
        sub: 'auth0|789',
      }
      expect(isNuonSession(user)).toBe(false)
    })
  })
})