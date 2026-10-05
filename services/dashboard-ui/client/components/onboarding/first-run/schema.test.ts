import { expect, test } from 'bun:test'
import { deploySchema, startSchema } from './schema'

test('start schema rejects a blank name, a spaced name, and a missing cloud', () => {
  expect(startSchema.safeParse({ appName: '', cloud: '' }).success).toBe(false)
  expect(startSchema.safeParse({ appName: 'My App', cloud: 'aws' }).success).toBe(false)
  expect(startSchema.safeParse({ appName: 'acme-api', cloud: '' }).success).toBe(false)
  expect(startSchema.safeParse({ appName: 'acme-api', cloud: 'aws' }).success).toBe(true)
})

test('deploy schema requires a region for the chosen cloud', () => {
  expect(deploySchema('aws').safeParse({ region: 'us-east-1', autoApprove: true }).success).toBe(true)
  expect(deploySchema('aws').safeParse({ region: 'eastus', autoApprove: true }).success).toBe(false)
  expect(deploySchema('azure').safeParse({ region: 'eastus', autoApprove: false }).success).toBe(true)
})
