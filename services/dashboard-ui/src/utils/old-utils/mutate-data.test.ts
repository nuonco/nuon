import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import type { TOrg } from '../../types'
import { mutateData, nueMutateData } from './mutate-data'

test.skip('mutateData should return an new org object when provided the orgs path', async () => {
  const spec = await mutateData<TOrg>({
    data: { name: 'test' },
    path: 'orgs',
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty('name')
})

test.skip('mutateData should throw an error with the default error message when it can not create an orgs', async () => {
  try {
    await mutateData<TOrg>({
      data: { name: 'test' },
      path: 'orgs',
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: Encountered an issue retrieving this information, please refresh the page to try again.]`
    )
  }
})

test.skip('mutateData should throw an error with a custom error message when it can not create an orgs', async () => {
  try {
    await mutateData<TOrg>({
      data: { name: 'test' },
      errorMessage: 'Custom error message!',
      path: 'orgs',
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Custom error message!]`)
  }
})

test.skip('nueMutateData should return a new org when POST to orgs path', async () => {
  const { data, error, status } = await nueMutateData<TOrg>({
    path: 'orgs',
    body: { name: 'test' },
  })

  expect(data).toHaveProperty('id')
  expect(data).toHaveProperty('created_at')
  expect(data).toHaveProperty('updated_at')
  expect(data).toHaveProperty('name')
  expect(status).toBe(201)
  expect(error).toBeNull()
})

test.skip('nueMutateData should return an 400 error when provided orgs path', async () => {
  const { data, error, status } = await nueMutateData<TOrg>({
    path: 'orgs',
    body: { name: 'test' },
  })

  expect(data).toBeNull()
  expect(status).toBe(400)
  expect(error).toHaveProperty('error')
  expect(error).toHaveProperty('description')
  expect(error).toHaveProperty('user_error')
})

test.skip('nueMutateData should return an 401 error when provided orgs path', async () => {
  const { data, error, status } = await nueMutateData<TOrg>({
    path: 'orgs',
    body: { name: 'test' },
  })

  expect(data).toBeNull()
  expect(status).toBe(401)
  expect(error).toHaveProperty('error')
  expect(error).toHaveProperty('description')
  expect(error).toHaveProperty('user_error')
})

test.skip('nueMutateData should return an 403 error when provided orgs path', async () => {
  const { data, error, status } = await nueMutateData<TOrg>({
    path: 'orgs',
    body: { name: 'test' },
  })

  expect(data).toBeNull()
  expect(status).toBe(403)
  expect(error).toHaveProperty('error')
  expect(error).toHaveProperty('description')
  expect(error).toHaveProperty('user_error')
})

test.skip('nueMutateData should return an 404 error when provided orgs path', async () => {
  const { data, error, status } = await nueMutateData<TOrg>({
    path: 'orgs',
    body: { name: 'test' },
  })

  expect(data).toBeNull()
  expect(status).toBe(404)
  expect(error).toHaveProperty('error')
  expect(error).toHaveProperty('description')
  expect(error).toHaveProperty('user_error')
})

test.skip('nueMutateData should return an 500 error when provided orgs path', async () => {
  const { data, error, status } = await nueMutateData<TOrg>({
    path: 'orgs',
    body: { name: 'test' },
  })

  expect(data).toBeNull()
  expect(status).toBe(500)
  expect(error).toHaveProperty('error')
  expect(error).toHaveProperty('description')
  expect(error).toHaveProperty('user_error')
})
