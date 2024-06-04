import '@test/mock-fetch-options'
import {  expect, test } from 'vitest'
import { getBuildLogs } from './get-build-logs'

const buildId = 'build-id'
const componentId = 'component-id'
const orgId = 'org-id'

test('getBuildLogs should return a build object', async () => {
  const spec = await getBuildLogs({
    componentId,
    buildId,
    orgId,
  })

  expect(spec).toMatchInlineSnapshot(`
    [
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
    ]
  `)  
})

test('getBuildLogs should throw an error when it can not find a build', async () => {
  try {
    await getBuildLogs({
      componentId,
      buildId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch build logs]`)
  }
  
})
