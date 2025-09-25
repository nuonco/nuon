import { describe, expect, test, vi, beforeEach } from 'vitest'
import { setOrgCookie, getOrgIdFromCookie } from './org-session-cookie'

// Mock Next.js cookies
const mockSet = vi.fn()
const mockGet = vi.fn()

vi.mock('next/headers', () => ({
  cookies: vi.fn(() => Promise.resolve({
    set: mockSet,
    get: mockGet,
  })),
}))

describe('org-session-cookie', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('setOrgCookie', () => {
    test('should set org_session cookie with provided orgId', async () => {
      const orgId = 'org-123'
      
      await setOrgCookie(orgId)

      expect(mockSet).toHaveBeenCalledWith('org_session', orgId, {
        path: '/',
        httpOnly: false,
        maxAge: 60 * 60 * 24 * 365,
        sameSite: 'lax',
      })
    })

    test('should set cookie with correct options', async () => {
      const orgId = 'org-456'
      
      await setOrgCookie(orgId)

      const expectedOptions = {
        path: '/',
        httpOnly: false,
        maxAge: 31536000, // 1 year in seconds
        sameSite: 'lax',
      }

      expect(mockSet).toHaveBeenCalledWith('org_session', orgId, expectedOptions)
    })

    test('should handle empty string orgId', async () => {
      const orgId = ''
      
      await setOrgCookie(orgId)

      expect(mockSet).toHaveBeenCalledWith('org_session', '', {
        path: '/',
        httpOnly: false,
        maxAge: 60 * 60 * 24 * 365,
        sameSite: 'lax',
      })
    })

    test('should handle special characters in orgId', async () => {
      const orgId = 'org-123-test_special'
      
      await setOrgCookie(orgId)

      expect(mockSet).toHaveBeenCalledWith('org_session', orgId, {
        path: '/',
        httpOnly: false,
        maxAge: 60 * 60 * 24 * 365,
        sameSite: 'lax',
      })
    })
  })

  describe('getOrgIdFromCookie', () => {
    test('should return orgId when cookie exists', async () => {
      const expectedOrgId = 'org-789'
      mockGet.mockReturnValue({ value: expectedOrgId })

      const result = await getOrgIdFromCookie()

      expect(mockGet).toHaveBeenCalledWith('org_session')
      expect(result).toBe(expectedOrgId)
    })

    test('should return undefined when cookie does not exist', async () => {
      mockGet.mockReturnValue(undefined)

      const result = await getOrgIdFromCookie()

      expect(mockGet).toHaveBeenCalledWith('org_session')
      expect(result).toBeUndefined()
    })

    test('should return undefined when cookie exists but has no value', async () => {
      mockGet.mockReturnValue({})

      const result = await getOrgIdFromCookie()

      expect(mockGet).toHaveBeenCalledWith('org_session')
      expect(result).toBeUndefined()
    })

    test('should return empty string if that is the cookie value', async () => {
      mockGet.mockReturnValue({ value: '' })

      const result = await getOrgIdFromCookie()

      expect(mockGet).toHaveBeenCalledWith('org_session')
      expect(result).toBe('')
    })

    test('should return null if that is the cookie value', async () => {
      mockGet.mockReturnValue({ value: null })

      const result = await getOrgIdFromCookie()

      expect(mockGet).toHaveBeenCalledWith('org_session')
      expect(result).toBeNull()
    })

    test('should handle complex orgId values', async () => {
      const complexOrgId = 'org-123-456_test-special'
      mockGet.mockReturnValue({ value: complexOrgId })

      const result = await getOrgIdFromCookie()

      expect(mockGet).toHaveBeenCalledWith('org_session')
      expect(result).toBe(complexOrgId)
    })
  })
})