import { describe, expect, test } from 'bun:test'
import { normalizeDiffOperation, serializeTerraform, terraformDiff } from '.'

describe('normalizeDiffOperation', () => {
  test('normalizes provider action vocabularies', () => {
    expect(normalizeDiffOperation('added')).toBe('create')
    expect(normalizeDiffOperation('modified')).toBe('update')
    expect(normalizeDiffOperation('create-replacement')).toBe('replace')
    expect(normalizeDiffOperation('destroyed')).toBe('delete')
    expect(normalizeDiffOperation('refresh')).toBe('read')
    expect(normalizeDiffOperation('same')).toBe('no-op')
  })

  test('rejects unknown operations', () => {
    expect(normalizeDiffOperation('mystery')).toBeUndefined()
  })
})

describe('serializeTerraform', () => {
  test('sorts every object deterministically', () => {
    const value = {
      zeta: { second: 2, first: 1 },
      alpha: true,
    }

    const expected = `alpha = true
zeta = {
  first = 1
  second = 2
}`

    expect(serializeTerraform(value)).toBe(expected)
    expect(serializeTerraform(value)).toBe(expected)
  })

  test('renders lists and nested maps as HCL-ish values', () => {
    expect(
      serializeTerraform({
        rules: [{ port: 443, protocol: 'tcp' }],
        tags: { environment: 'production' },
      })
    ).toBe(`rules = [
  {
    port = 443
    protocol = "tcp"
  },
]
tags = {
  environment = "production"
}`)
  })

  test('applies unknown and sensitive masks as values', () => {
    const diff = terraformDiff({
      before: { endpoint: null, token: 'old' },
      after: { endpoint: null, token: 'new' },
      afterUnknown: { endpoint: true },
      beforeSensitive: { token: true },
      afterSensitive: { token: true },
    })

    expect(diff.before).toContain('token = (sensitive value)')
    expect(diff.after).toContain('endpoint = (known after apply)')
    expect(diff.after).toContain('token = (sensitive value)')
  })

  test('does not invent a difference from key order', () => {
    const diff = terraformDiff({
      before: { beta: 2, alpha: { zeta: 3, gamma: 1 } },
      after: { alpha: { gamma: 1, zeta: 3 }, beta: 2 },
    })

    expect(diff.before).toBe(diff.after)
  })
})
