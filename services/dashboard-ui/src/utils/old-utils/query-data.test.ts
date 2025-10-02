import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import type { TApp } from '../types'
import { nueQueryData } from './query-data'

const orgId = 'org-id'

test.skip('nueQueryData should return a list of apps when provided apps path', async () => {
  const { data, error, status } = await nueQueryData<Array<TApp>>({
    path: 'apps',
    orgId,
  })

  expect(data).toHaveLength(2)
  data.map((d) => {
    expect(d).toHaveProperty('id')
    expect(d).toHaveProperty('created_at')
    expect(d).toHaveProperty('updated_at')
    expect(d).toHaveProperty('name')
    expect(d).toHaveProperty('cloud_platform')
  })
  expect(status).toBe(200)
  expect(error).toBeNull()
})

test.skip('nueQueryData should return an 400 error when provided apps path', async () => {
  const { data, error, status } = await nueQueryData<Array<TApp>>({
    path: 'apps',
    orgId,
  })

  expect(data).toBeNull()
  expect(status).toBe(400)
  expect(error).toHaveProperty('error')
  expect(error).toHaveProperty('description')
  expect(error).toHaveProperty('user_error')
})

test.skip('nueQueryData should return an 401 error when provided apps path', async () => {
  const { data, error, status } = await nueQueryData<Array<TApp>>({
    path: 'apps',
    orgId,
  })

  expect(data).toBeNull()
  expect(status).toBe(401)
  expect(error).toHaveProperty('error')
  expect(error).toHaveProperty('description')
  expect(error).toHaveProperty('user_error')
})

test.skip('nueQueryData should return an 403 error when provided apps path', async () => {
  const { data, error, status } = await nueQueryData<Array<TApp>>({
    path: 'apps',
    orgId,
  })

  expect(data).toBeNull()
  expect(status).toBe(403)
  expect(error).toHaveProperty('error')
  expect(error).toHaveProperty('description')
  expect(error).toHaveProperty('user_error')
})

test.skip('nueQueryData should return an 404 error when provided apps path', async () => {
  const { data, error, status } = await nueQueryData<Array<TApp>>({
    path: 'apps',
    orgId,
  })

  expect(data).toBeNull()
  expect(status).toBe(404)
  expect(error).toHaveProperty('error')
  expect(error).toHaveProperty('description')
  expect(error).toHaveProperty('user_error')
})

test.skip('nueQueryData should return an 500 error when provided apps path', async () => {
  const { data, error, status } = await nueQueryData<Array<TApp>>({
    path: 'apps',
    orgId,
  })

  expect(data).toBeNull()
  expect(status).toBe(500)
  expect(error).toHaveProperty('error')
  expect(error).toHaveProperty('description')
  expect(error).toHaveProperty('user_error')
})
