import '@test/mock-fetch-options'
import { expect, test, } from 'vitest'
import type { TApp } from "../types"
import { queryData } from './query-data'

const orgId = 'org-id'


test('queryData should return a list of apps when provided apps path', async () => {
  const spec = await queryData<Array<TApp>>({
    path: "apps",
    orgId,
  })

  expect(spec).toHaveLength(3)
  spec.map((s) => {
    expect(s).toHaveProperty('id')
    expect(s).toHaveProperty('created_at')
    expect(s).toHaveProperty('updated_at')
    expect(s).toHaveProperty('name')
    expect(s).toHaveProperty('cloud_platform')
  })
})

test('queryData should throw an error with the default error message when it can not find apps', async () => {
  try {
    await queryData<Array<TApp>>({
      path: "apps",
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Encountered an issue retrieving this information, please refresh the page to try again.]`)
  }
})


test('queryData should throw an error with a custom error message when it can not find apps', async () => {
  try {
    await queryData<Array<TApp>>({
      errorMessage: "Custom error message!",
      path: "apps",
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Custom error message!]`)
  }
})
