import { describe, expect, test } from 'bun:test'
import { inviteMemberSchema, inviteMemberValuesValid } from './InviteMember'

describe('InviteMember', () => {
  test('rejects an invalid email', () => {
    expect(
      inviteMemberValuesValid({
        email: 'not-an-email',
        roleType: 'org_read_only',
      })
    ).toBe(false)
  })

  test('keeps submission blocked without a role', () => {
    expect(
      inviteMemberValuesValid({
        email: 'member@example.com',
        roleType: '',
      })
    ).toBe(false)
  })

  test('accepts a valid flat form value', () => {
    expect(
      inviteMemberValuesValid({
        email: 'member@example.com',
        roleType: 'org_read_only',
      })
    ).toBe(true)
  })

  test('normalizes surrounding email whitespace', () => {
    expect(
      inviteMemberSchema.parse({
        email: '  member@example.com  ',
        roleType: 'org_read_only',
      }).email
    ).toBe('member@example.com')
  })
})
