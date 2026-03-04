import { expect, test } from 'vitest'
import type { TApp } from '../types'
import { api } from './api'

const orgId = 'org-id'

test('api should return a list of apps when provided apps path', async () => {
  const result = await api<TApp[]>({ path: 'apps', orgId })
  expect(Array.isArray(result)).toBe(true)
})

test.each([[400], [401], [403], [404], [500]])('%s status rejects', async () => {
  await expect(api<TApp[]>({ path: 'apps', orgId })).rejects.toMatchObject({
    error: expect.any(String),
    description: expect.any(String),
    user_error: expect.any(Boolean),
  })
})
