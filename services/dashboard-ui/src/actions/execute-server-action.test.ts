import { describe, expect, test, vi, beforeEach } from 'vitest'
import { executeServerAction } from './execute-server-action'

// Mock Next.js revalidatePath
vi.mock('next/cache', () => ({
  revalidatePath: vi.fn(),
}))

describe('executeServerAction', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  test('should execute action and return result without path revalidation', async () => {
    const mockAction = vi.fn().mockResolvedValue({ success: true, data: 'test-data' })
    const args = { orgId: 'org-123' }

    const result = await executeServerAction({
      action: mockAction,
      args,
    })

    expect(mockAction).toHaveBeenCalledWith(args)
    expect(result).toEqual({ success: true, data: 'test-data' })
  })

  test('should execute action and revalidate path when path is provided', async () => {
    const { revalidatePath } = await import('next/cache')
    const mockAction = vi.fn().mockResolvedValue({ id: 'test-id', name: 'test-name' })
    const args = { orgId: 'org-456', name: 'Test App' }
    const path = '/org-456/apps'

    const result = await executeServerAction({
      action: mockAction,
      args,
      path,
    })

    expect(mockAction).toHaveBeenCalledWith(args)
    expect(revalidatePath).toHaveBeenCalledWith(path)
    expect(result).toEqual({ id: 'test-id', name: 'test-name' })
  })

  test('should handle action that throws an error', async () => {
    const mockError = new Error('Action failed')
    const mockAction = vi.fn().mockRejectedValue(mockError)
    const args = { orgId: 'org-789' }

    await expect(executeServerAction({
      action: mockAction,
      args,
    })).rejects.toThrow('Action failed')

    expect(mockAction).toHaveBeenCalledWith(args)
  })

  test('should handle action that throws an error with path provided', async () => {
    const { revalidatePath } = await import('next/cache')
    const mockError = new Error('Action with path failed')
    const mockAction = vi.fn().mockRejectedValue(mockError)
    const args = { orgId: 'org-error' }
    const path = '/org-error/apps'

    await expect(executeServerAction({
      action: mockAction,
      args,
      path,
    })).rejects.toThrow('Action with path failed')

    expect(mockAction).toHaveBeenCalledWith(args)
    expect(revalidatePath).not.toHaveBeenCalled()
  })

  test('should work with different argument types', async () => {
    const mockAction = vi.fn().mockResolvedValue('string result')
    const stringArgs = 'simple-string'

    const result = await executeServerAction({
      action: mockAction,
      args: stringArgs,
    })

    expect(mockAction).toHaveBeenCalledWith(stringArgs)
    expect(result).toBe('string result')
  })

  test('should work with complex nested argument objects', async () => {
    const mockAction = vi.fn().mockResolvedValue({ processed: true })
    const complexArgs = {
      orgId: 'org-complex',
      config: {
        settings: {
          enabled: true,
          values: [1, 2, 3],
        },
        metadata: {
          version: '1.0.0',
          author: 'test-user',
        },
      },
    }

    const result = await executeServerAction({
      action: mockAction,
      args: complexArgs,
    })

    expect(mockAction).toHaveBeenCalledWith(complexArgs)
    expect(result).toEqual({ processed: true })
  })

  test('should work with undefined/null return values', async () => {
    const mockAction = vi.fn().mockResolvedValue(undefined)
    const args = { orgId: 'org-undefined' }

    const result = await executeServerAction({
      action: mockAction,
      args,
    })

    expect(mockAction).toHaveBeenCalledWith(args)
    expect(result).toBeUndefined()
  })

  test('should revalidate path even when action returns undefined', async () => {
    const { revalidatePath } = await import('next/cache')
    const mockAction = vi.fn().mockResolvedValue(undefined)
    const args = { orgId: 'org-undefined' }
    const path = '/org-undefined/dashboard'

    const result = await executeServerAction({
      action: mockAction,
      args,
      path,
    })

    expect(mockAction).toHaveBeenCalledWith(args)
    expect(revalidatePath).toHaveBeenCalledWith(path)
    expect(result).toBeUndefined()
  })
})