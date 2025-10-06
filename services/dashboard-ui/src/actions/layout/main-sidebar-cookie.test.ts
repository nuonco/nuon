import { describe, expect, test, vi, beforeEach } from 'vitest'
import {
  setSidebarCookie,
  getIsSidebarOpenFromCookie,
} from './main-sidebar-cookie'

// Mock Next.js cookies
const mockSet = vi.fn()
const mockGet = vi.fn()

vi.mock('next/headers', () => ({
  cookies: vi.fn(() =>
    Promise.resolve({
      set: mockSet,
      get: mockGet,
    })
  ),
}))

describe('main-sidebar-cookie', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('setSidebarCookie', () => {
    test('should set cookie to "1" when isOpen is true', async () => {
      await setSidebarCookie(true)

      expect(mockSet).toHaveBeenCalledWith('sidebar_open', '1', {
        path: '/',
        httpOnly: false,
        maxAge: 60 * 60 * 24 * 365,
        sameSite: 'lax',
      })
    })

    test('should set cookie to "0" when isOpen is false', async () => {
      await setSidebarCookie(false)

      expect(mockSet).toHaveBeenCalledWith('sidebar_open', '0', {
        path: '/',
        httpOnly: false,
        maxAge: 60 * 60 * 24 * 365,
        sameSite: 'lax',
      })
    })

    test('should set cookie with correct options', async () => {
      await setSidebarCookie(true)

      const expectedOptions = {
        path: '/',
        httpOnly: false,
        maxAge: 31536000, // 1 year in seconds
        sameSite: 'lax',
      }

      expect(mockSet).toHaveBeenCalledWith('sidebar_open', '1', expectedOptions)
    })
  })

  describe('getIsSidebarOpenFromCookie', () => {
    test('should return true when cookie value is "1"', async () => {
      mockGet.mockReturnValue({ value: '1' })

      const result = await getIsSidebarOpenFromCookie()

      expect(mockGet).toHaveBeenCalledWith('sidebar_open')
      expect(result).toBe(true)
    })

    test('should return false when cookie value is "0"', async () => {
      mockGet.mockReturnValue({ value: '0' })

      const result = await getIsSidebarOpenFromCookie()

      expect(mockGet).toHaveBeenCalledWith('sidebar_open')
      expect(result).toBe(false)
    })

    test('should return false when cookie value is something else', async () => {
      mockGet.mockReturnValue({ value: 'random' })

      const result = await getIsSidebarOpenFromCookie()

      expect(mockGet).toHaveBeenCalledWith('sidebar_open')
      expect(result).toBe(false)
    })

    test('should return false when cookie does not exist', async () => {
      mockGet.mockReturnValue(undefined)

      const result = await getIsSidebarOpenFromCookie()

      expect(mockGet).toHaveBeenCalledWith('sidebar_open')
      expect(result).toBe(false)
    })

    test('should return false when cookie exists but has no value', async () => {
      mockGet.mockReturnValue({})

      const result = await getIsSidebarOpenFromCookie()

      expect(mockGet).toHaveBeenCalledWith('sidebar_open')
      expect(result).toBe(false)
    })
  })
})
