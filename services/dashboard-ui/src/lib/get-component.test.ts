import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getComponent } from './get-component'

const componentId = 'component-id'
const orgId = 'org-id'

test('getComponent should return a component object', async () => {
  const spec = await getComponent({
    componentId,
    orgId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty('name')
})

test('getComponent should throw an error when it can not find a component', async () => {
  try {
    await getComponent({
      componentId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch component]`)
  }
})
