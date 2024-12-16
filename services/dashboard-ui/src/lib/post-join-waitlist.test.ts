import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { postJoinWaitlist } from './post-join-waitlist'


const orgName = 'test-org'


test('postJoinWaitlist should return an new waitlist object', async () => {
  const spec = await postJoinWaitlist({ org_name: orgName })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_by_id')
  expect(spec).toHaveProperty('created_at') 
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty("org_name")
})

test.skip('postJoinWaitlist should throw an error when it can not create a waitlist object', async () => {
  try {
    await postJoinWaitlist({ org_name: orgName })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error]`)
  }
})
