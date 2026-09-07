import { describe, expect, test } from 'bun:test'
import {
  booleanQueryParameter,
  commaSetQueryParameter,
  enumQueryParameter,
  offsetQueryParameter,
  readQueryParameter,
  repeatedSetQueryParameter,
  stringQueryParameter,
  writeQueryParameter,
} from './list-query'

describe('list query codecs', () => {
  test('reads and writes strings while omitting empty values', () => {
    const parameter = stringQueryParameter('q')
    const params = new URLSearchParams('q=payments')

    expect(readQueryParameter(params, parameter)).toBe('payments')

    writeQueryParameter(params, parameter, '')
    expect(params.has('q')).toBe(false)
  })

  test('falls back for invalid booleans and enums', () => {
    const booleanParameter = booleanQueryParameter('active', {
      defaultValue: true,
    })
    const enumParameter = enumQueryParameter(
      'status',
      ['all', 'active', 'failed'] as const,
      { defaultValue: 'all' }
    )
    const params = new URLSearchParams('active=sometimes&status=unknown')

    expect(readQueryParameter(params, booleanParameter)).toBe(true)
    expect(readQueryParameter(params, enumParameter)).toBe('all')
  })

  test('omits declared defaults', () => {
    const booleanParameter = booleanQueryParameter('active', {
      defaultValue: true,
    })
    const enumParameter = enumQueryParameter(
      'status',
      ['all', 'active'] as const,
      { defaultValue: 'all' }
    )
    const params = new URLSearchParams('active=false&status=active')

    writeQueryParameter(params, booleanParameter, true)
    writeQueryParameter(params, enumParameter, 'all')

    expect(params.toString()).toBe('')
  })

  test('canonicalizes comma-separated sets', () => {
    const parameter = commaSetQueryParameter<'helm' | 'terraform'>('types')
    const params = new URLSearchParams('types=terraform,helm,terraform')

    const value = readQueryParameter(params, parameter)
    expect([...value].sort()).toEqual(['helm', 'terraform'])

    writeQueryParameter(params, parameter, value)
    expect(params.toString()).toBe('types=helm%2Cterraform')
    expect(parameter.codec.toQueryKey(value)).toEqual(['helm', 'terraform'])
  })

  test('round trips repeated sets in deterministic order', () => {
    const parameter = repeatedSetQueryParameter('severity')
    const params = new URLSearchParams('severity=warn&severity=info')
    const value = readQueryParameter(params, parameter)

    writeQueryParameter(params, parameter, value)

    expect(params.getAll('severity')).toEqual(['info', 'warn'])
    expect(readQueryParameter(params, parameter)).toEqual(
      new Set(['info', 'warn'])
    )
  })

  test('uses fresh sets for declared defaults', () => {
    const parameter = commaSetQueryParameter('types', {
      defaultValue: ['helm', 'terraform'],
    })
    const first = readQueryParameter(new URLSearchParams(), parameter)
    first.delete('helm')

    expect([...readQueryParameter(new URLSearchParams(), parameter)]).toEqual([
      'helm',
      'terraform',
    ])
  })

  test('normalizes invalid and fractional offsets', () => {
    const parameter = offsetQueryParameter()

    expect(
      readQueryParameter(new URLSearchParams('offset=-1'), parameter)
    ).toBe(0)
    expect(
      readQueryParameter(new URLSearchParams('offset=nope'), parameter)
    ).toBe(0)
    expect(
      readQueryParameter(new URLSearchParams('offset=21.9'), parameter)
    ).toBe(21)

    const params = new URLSearchParams('offset=20')
    writeQueryParameter(params, parameter, 0)
    expect(params.toString()).toBe('')
  })
})
