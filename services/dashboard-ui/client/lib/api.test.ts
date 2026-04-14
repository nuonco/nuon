import { expect, test } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@test/mock-api-server'
import type { TApp } from '../types'
import { api } from './api'

const orgId = 'org-id'
const baseURL = 'http://localhost:8081'

test('api should return a list of apps when provided apps path', async () => {
  const result = await api<TApp[]>({ path: 'apps', orgId })
  expect(Array.isArray(result)).toBe(true)
})

test.each([[400], [401], [403], [404], [500]])(
  '%s status rejects',
  async (status) => {
    server.use(
      http.get(`${baseURL}/v1/apps`, () => {
        return HttpResponse.json(
          {
            error: 'test error',
            description: 'test description',
            user_error: false,
          },
          { status }
        )
      })
    )

    if (status === 401) {
      await expect(
        api<TApp[]>({ path: 'apps', orgId })
      ).resolves.toBeUndefined()
      expect(window.location.reload).toHaveBeenCalled()
    } else {
      await expect(
        api<TApp[]>({ path: 'apps', orgId })
      ).rejects.toMatchObject({
        error: expect.any(String),
        description: expect.any(String),
        user_error: expect.any(Boolean),
      })
    }
  }
)
