import { expect, test, mock, beforeAll, afterEach, afterAll } from 'bun:test'
import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import type { TApp } from '../types'
import { api } from './api'

const orgId = 'org-id'
const baseUrl = 'http://test.local'

const server = setupServer(
  http.get(`${baseUrl}/v1/apps`, () => {
    return HttpResponse.json([], { status: 200 })
  })
)

Object.defineProperty(window, 'location', {
  value: { ...window.location, reload: mock() },
})

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

test('api should return a list of apps when provided apps path', async () => {
  const result = await api<TApp[]>({ path: 'apps', orgId, baseUrl })
  expect(Array.isArray(result)).toBe(true)
})

test.each([[400], [401], [403], [404], [500]])(
  '%s status rejects',
  async (status) => {
    server.use(
      http.get(`${baseUrl}/v1/apps`, () => {
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
        api<TApp[]>({ path: 'apps', orgId, baseUrl })
      ).resolves.toBeUndefined()
      expect(window.location.reload).toHaveBeenCalled()
    } else {
      await expect(
        api<TApp[]>({ path: 'apps', orgId, baseUrl })
      ).rejects.toMatchObject({
        error: expect.any(String),
        description: expect.any(String),
        user_error: expect.any(Boolean),
      })
    }
  }
)
