import { afterEach, beforeEach, expect, test } from 'bun:test'
import { createTrialOrg } from './create-org'

type TCall = { url: string; method: string; body?: any }

const calls: TCall[] = []
let names: string[] = []
let orgStatuses: number[] = []

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'content-type': 'application/json' },
  })

const realFetch = globalThis.fetch

beforeEach(() => {
  calls.length = 0
  names = ['taken-name-one', 'free-name-two', 'free-name-three']
  orgStatuses = []
  globalThis.fetch = Object.assign(
    async (input: Parameters<typeof fetch>[0], init?: RequestInit) => {
      const url = String(input)
      const method = init?.method ?? 'GET'
      calls.push({ url, method, body: init?.body ? JSON.parse(String(init.body)) : undefined })
      if (url === '/api/random-name') return json(200, { name: names.shift() })
      if (url === '/v1/orgs' && method === 'POST') {
        const status = orgStatuses.shift() ?? 201
        if (status === 201) return json(201, { id: 'orgabc123', name: calls.at(-1)?.body?.name })
        return json(status, { error: 'conflict', description: 'name taken', user_error: true })
      }
      return json(404, { error: 'not found' })
    },
    { preconnect: () => {} }
  ) as unknown as typeof fetch
})

afterEach(() => {
  globalThis.fetch = realFetch
})

const orgCreates = () => calls.filter((c) => c.url === '/v1/orgs' && c.method === 'POST')

test('retries with a new name after a 409 and creates one org', async () => {
  orgStatuses = [409, 201]

  const org = await createTrialOrg()

  expect(org.id).toBe('orgabc123')
  expect(orgCreates().map((c) => c.body.name)).toEqual(['taken-name-one', 'free-name-two'])
  expect(orgCreates()[1].body).toEqual({
    name: 'free-name-two',
    use_sandbox_mode: false,
    tags: ['Trial'],
  })
})

test('gives up after the attempt limit', async () => {
  names = Array.from({ length: 6 }, (_, i) => `name-${i}`)
  orgStatuses = [409, 409, 409, 409, 409, 409]

  await expect(createTrialOrg()).rejects.toMatchObject({ status: 409 })
  expect(orgCreates()).toHaveLength(5)
})

test('does not retry other errors', async () => {
  orgStatuses = [500]

  await expect(createTrialOrg()).rejects.toMatchObject({ status: 500 })
  expect(orgCreates()).toHaveLength(1)
})
