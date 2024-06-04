import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getBuildPlan } from './get-build-plan'

const buildId = 'build-id'
const componentId = 'component-id'
const orgId = 'org-id'

test('getBuildPlan should return a build object', async () => {
  const spec = await getBuildPlan({
    componentId,
    buildId,
    orgId,
  })

  expect(spec?.actual).toBeNull()
})

test('getBuildPlan should throw an error when it can not find a build', async () => {
  try {
    await getBuildPlan({
      componentId,
      buildId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch build plan]`)
  }
})
