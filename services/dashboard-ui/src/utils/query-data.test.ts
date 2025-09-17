import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import type { TApp } from '../types'
import { queryData, nueQueryData } from './query-data'

const orgId = 'org-id'

test.skip('queryData should return a list of apps when provided apps path', async () => {
  const spec = await queryData<Array<TApp>>({
    path: 'apps',
    orgId,
  })

  expect(spec).toHaveLength(2)
  spec.map((s) => {
    expect(s).toHaveProperty('id')
    expect(s).toHaveProperty('created_at')
    expect(s).toHaveProperty('updated_at')
    expect(s).toHaveProperty('name')
    expect(s).toHaveProperty('cloud_platform')
  })
})

test.skip('queryData should throw an error with the default error message when it can not find apps', async () => {
  try {
    await queryData<Array<TApp>>({
      path: 'apps',
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: Encountered an issue retrieving this information, please refresh the page to try again.]`
    )
  }
})

test.skip('queryData should throw an error with a custom error message when it can not find apps', async () => {
  try {
    await queryData<Array<TApp>>({
      errorMessage: 'Custom error message!',
      path: 'apps',
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Custom error message!]`)
  }
})

test('nueQueryData should return a list of apps when provided apps path', async () => {
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

test('nueQueryData should return an 400 error when provided apps path', async () => {
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

test('nueQueryData should return an 401 error when provided apps path', async () => {
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

test('nueQueryData should return an 403 error when provided apps path', async () => {
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

test('nueQueryData should return an 404 error when provided apps path', async () => {
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

test('nueQueryData should return an 500 error when provided apps path', async () => {
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
