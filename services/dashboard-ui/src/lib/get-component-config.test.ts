import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getComponentConfig } from './get-component-config'

const componentId = 'component-id'
const orgId = 'org-id'

test('getComponentConfig should return a config object', async () => {
  const spec = await getComponentConfig({
    componentId,
    orgId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty('version')
})

test('getComponentConfig should throw an error when it can not find a config', async () => {
  try {
    await getComponentConfig({
      componentId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: Failed to fetch component config]`
    )
  }
})
