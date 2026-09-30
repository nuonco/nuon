import { expect, test } from 'bun:test'
import type { TAPIError } from '@/types'
import { orgNameSubmitError } from './org-name-error'

const conflict = {
  error:
    'unable to update org: ERROR: duplicate key value violates unique constraint "idx_org_name" (SQLSTATE 23505)',
  description: 'duplicate key',
  user_error: true,
  status: 409,
} as TAPIError

test('rewrites a unique-name conflict into a name message', () => {
  expect(orgNameSubmitError(conflict)).toEqual({
    ...conflict,
    error: 'An organization with this name already exists.',
    description: 'Choose a different name.',
  })
})

test('rewrites a duplicate-key failure even when the status is not 409', () => {
  const error = { ...conflict, status: 500 }
  expect(orgNameSubmitError(error)?.error).toBe(
    'An organization with this name already exists.'
  )
})

test('leaves permission and validation errors unchanged', () => {
  const forbidden = {
    error: 'this action requires write access to organization settings',
    description:
      'Your role (Read-only) does not have write access to organization settings. Ask an organization admin to assign a role that does.',
    user_error: true,
    status: 403,
  } as TAPIError
  expect(orgNameSubmitError(forbidden)).toBe(forbidden)
  expect(orgNameSubmitError(null)).toBeNull()
})
