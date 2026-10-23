import { expect, test } from 'bun:test'
import { orgNameFromEmail } from './org-name-from-email'

test('names the org after a work domain', () => {
  expect(orgNameFromEmail('jane@acme.com')).toBe('Acme')
  expect(orgNameFromEmail('jane@acme.io')).toBe('Acme')
  expect(orgNameFromEmail('Jane@ACME-Labs.dev')).toBe('Acme-labs')
  expect(orgNameFromEmail('jane@eng.acme.com')).toBe('Acme')
  expect(orgNameFromEmail('jane@acme.co.uk')).toBe('Acme')
  expect(orgNameFromEmail('jane@acme.com.au')).toBe('Acme')
  expect(orgNameFromEmail('jane@acme.co')).toBe('Acme')
})

test('falls back for free email providers', () => {
  for (const email of [
    'jane@gmail.com',
    'jane@yahoo.com',
    'jane@yahoo.co.uk',
    'jane@icloud.com',
    'jane@me.com',
    'jane@hotmail.fr',
    'jane@outlook.com',
    'jane@proton.me',
    'jane@protonmail.com',
    'jane@gmx.de',
    'jane@aol.com',
  ]) {
    expect(orgNameFromEmail(email)).toBeUndefined()
  }
})

test('falls back when the email does not parse', () => {
  expect(orgNameFromEmail(undefined)).toBeUndefined()
  expect(orgNameFromEmail('')).toBeUndefined()
  expect(orgNameFromEmail('jane')).toBeUndefined()
  expect(orgNameFromEmail('jane@localhost')).toBeUndefined()
})
