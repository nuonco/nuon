import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAccountsByOrgId } from './get-accounts-by-org-id'

describe('getAccountsByOrgId should handle response status codes from GET endpoint', () => {
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data } = await getAccountsByOrgId({
      orgId,
      limit: 10,
      offset: 0,
    })
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('name')
    expect(data).toHaveProperty('email')
    expect(data).toHaveProperty('created_at')
  })

  test('200 status without pagination params', async () => {
    const { data } = await getAccountsByOrgId({ orgId })
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('name')
    expect(data).toHaveProperty('email')
    expect(data).toHaveProperty('created_at')
  })

  test('200 status with only limit param', async () => {
    const { data } = await getAccountsByOrgId({
      orgId,
      limit: 5,
    })
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('name')
    expect(data).toHaveProperty('email')
    expect(data).toHaveProperty('created_at')
  })

  test('200 status with only offset param', async () => {
    const { data } = await getAccountsByOrgId({
      orgId,
      offset: 20,
    })
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('name')
    expect(data).toHaveProperty('email')
    expect(data).toHaveProperty('created_at')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getAccountsByOrgId({
      orgId,
      limit: 10,
      offset: 0,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })

  test.each(badResponseCodes)('%s status without pagination', async (code) => {
    const { error, status } = await getAccountsByOrgId({ orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})