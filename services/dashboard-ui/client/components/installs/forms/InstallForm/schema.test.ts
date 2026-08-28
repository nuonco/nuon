import { describe, expect, test } from 'bun:test'
import type { TAppInputConfig, TInstall } from '@/types'
import { buildInstallSchema, getEditableInputs, isBooleanInput } from './schema'
import { buildInstallDefaults, mergeDraftValues } from './defaults'

const inputConfig = {
  id: 'cfg-1',
  input_groups: [
    {
      id: 'g1',
      app_inputs: [
        {
          id: 'i1',
          name: 'db_url',
          type: 'string',
          required: true,
          source: 'vendor',
        },
        {
          id: 'i2',
          name: 'replicas',
          type: 'number',
          default: '1',
          source: 'vendor',
        },
        {
          id: 'i3',
          name: 'enable_tls',
          type: 'bool',
          default: 'false',
          source: 'vendor',
        },
        {
          id: 'i4',
          name: 'secret_key',
          type: 'string',
          required: true,
          source: 'customer',
        },
      ],
    },
  ],
} as unknown as TAppInputConfig

describe('isBooleanInput', () => {
  test('detects bool type', () => {
    expect(isBooleanInput({ type: 'bool' } as any)).toBe(true)
  })
  test('detects true/false defaults', () => {
    expect(isBooleanInput({ default: 'true' } as any)).toBe(true)
    expect(isBooleanInput({ default: 'false' } as any)).toBe(true)
  })
  test('rejects plain strings', () => {
    expect(isBooleanInput({ type: 'string', default: 'x' } as any)).toBe(false)
  })
})

describe('getEditableInputs', () => {
  test('excludes customer-source inputs', () => {
    const names = getEditableInputs(inputConfig).map((i) => i.name)
    expect(names).toEqual(['db_url', 'replicas', 'enable_tls'])
  })
  test('handles missing config', () => {
    expect(getEditableInputs(undefined)).toEqual([])
  })
})

describe('buildInstallSchema — create', () => {
  test('requires name and aws region', () => {
    const schema = buildInstallSchema({
      mode: 'create',
      platform: 'aws',
      inputConfig,
    })
    const res = schema.safeParse({
      name: '',
      region: '',
      inputs: { db_url: '', replicas: '1', enable_tls: false },
    })
    expect(res.success).toBe(false)
  })

  test('passes with valid create values', () => {
    const schema = buildInstallSchema({
      mode: 'create',
      platform: 'aws',
      inputConfig,
    })
    const res = schema.safeParse({
      name: 'my-install',
      region: 'us-west-2',
      aws_account_id: '',
      inputs: { db_url: 'postgres://x', replicas: '3', enable_tls: true },
    })
    expect(res.success).toBe(true)
  })

  test('required input must be non-empty', () => {
    const schema = buildInstallSchema({
      mode: 'create',
      platform: 'aws',
      inputConfig,
    })
    const res = schema.safeParse({
      name: 'my-install',
      region: 'us-west-2',
      inputs: { db_url: '', replicas: '1', enable_tls: false },
    })
    expect(res.success).toBe(false)
  })

  test('aws_account_id validated as 12 digits when required', () => {
    const schema = buildInstallSchema({
      mode: 'create',
      platform: 'aws',
      inputConfig,
      requireTargetAccount: true,
    })
    expect(
      schema.safeParse({
        name: 'x',
        region: 'us-west-2',
        aws_account_id: '123',
        inputs: { db_url: 'y', replicas: '1', enable_tls: false },
      }).success
    ).toBe(false)
    expect(
      schema.safeParse({
        name: 'x',
        region: 'us-west-2',
        aws_account_id: '123456789012',
        inputs: { db_url: 'y', replicas: '1', enable_tls: false },
      }).success
    ).toBe(true)
  })

  test('optional aws_account_id allows empty but rejects malformed', () => {
    const schema = buildInstallSchema({
      mode: 'create',
      platform: 'aws',
      inputConfig,
    })
    const base = {
      name: 'x',
      region: 'us-west-2',
      inputs: { db_url: 'y', replicas: '1', enable_tls: false },
    }
    expect(schema.safeParse({ ...base, aws_account_id: '' }).success).toBe(true)
    expect(schema.safeParse({ ...base, aws_account_id: '99' }).success).toBe(
      false
    )
  })
})

describe('buildInstallSchema — edit', () => {
  test('name optional when showNameField is false', () => {
    const schema = buildInstallSchema({ mode: 'edit', inputConfig })
    const res = schema.safeParse({
      name: '',
      inputs: { db_url: 'y', replicas: '1', enable_tls: false },
    })
    expect(res.success).toBe(true)
  })

  test('name required when showNameField is true', () => {
    const schema = buildInstallSchema({
      mode: 'edit',
      inputConfig,
      showNameField: true,
    })
    const res = schema.safeParse({
      name: '',
      inputs: { db_url: 'y', replicas: '1', enable_tls: false },
    })
    expect(res.success).toBe(false)
  })
})

describe('buildInstallDefaults', () => {
  test('derives typed input defaults from config', () => {
    const defaults = buildInstallDefaults({ mode: 'create', inputConfig })
    expect(defaults.inputs).toEqual({
      db_url: '',
      replicas: '1',
      enable_tls: false,
    })
    expect(defaults.name).toBe('')
    expect(defaults.deployDependents).toBe(true)
  })

  test('prefers install values in edit mode', () => {
    const install = {
      name: 'existing',
      install_inputs: [{ values: { db_url: 'live', enable_tls: 'true' } }],
    } as unknown as TInstall
    const defaults = buildInstallDefaults({
      mode: 'edit',
      inputConfig,
      install,
    })
    expect(defaults.name).toBe('existing')
    expect(defaults.inputs.db_url).toBe('live')
    expect(defaults.inputs.enable_tls).toBe(true)
    expect(defaults.inputs.replicas).toBe('1')
  })
})

describe('mergeDraftValues', () => {
  test('draft overrides base, inputs merge', () => {
    const base = buildInstallDefaults({ mode: 'create', inputConfig })
    const merged = mergeDraftValues(base, {
      name: 'drafted',
      inputs: { db_url: 'drafted-url' },
    })
    expect(merged.name).toBe('drafted')
    expect(merged.inputs.db_url).toBe('drafted-url')
    expect(merged.inputs.replicas).toBe('1')
  })

  test('null draft returns base unchanged', () => {
    const base = buildInstallDefaults({ mode: 'create', inputConfig })
    expect(mergeDraftValues(base, null)).toBe(base)
  })
})
