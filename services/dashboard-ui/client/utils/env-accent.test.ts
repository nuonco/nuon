import { describe, expect, test } from 'bun:test'
import { resolveEnvAccent } from './env-accent'

const org = {
  env_accent_config: {
    label_key: 'env',
    values: {
      production: 'error' as const,
      staging: 'warn' as const,
    },
  },
}

describe('resolveEnvAccent', () => {
  test('returns the configured accent for a matching install label', () => {
    const accent = resolveEnvAccent(
      { labels: { env: 'production' } },
      org,
    )
    expect(accent).toEqual({
      value: 'production',
      color: 'error',
      labelKey: 'env',
    })
  })

  test('returns null when the install has no labels', () => {
    expect(resolveEnvAccent({ labels: undefined }, org)).toBeNull()
  })

  test('returns null when the label value is not mapped', () => {
    expect(
      resolveEnvAccent({ labels: { env: 'unknown' } }, org),
    ).toBeNull()
  })

  test('returns null when the install lacks the configured label key', () => {
    expect(
      resolveEnvAccent({ labels: { tier: 'production' } }, org),
    ).toBeNull()
  })

  test('returns null when the org has no config', () => {
    expect(resolveEnvAccent({ labels: { env: 'production' } }, undefined)).toBeNull()
    expect(
      resolveEnvAccent({ labels: { env: 'production' } }, {
        env_accent_config: undefined,
      }),
    ).toBeNull()
  })

  test('returns null when label_key is empty', () => {
    expect(
      resolveEnvAccent(
        { labels: { env: 'production' } },
        { env_accent_config: { label_key: '', values: { production: 'error' } } },
      ),
    ).toBeNull()
  })

  test('honors a custom label key', () => {
    const accent = resolveEnvAccent(
      { labels: { tier: 'staging' } },
      {
        env_accent_config: {
          label_key: 'tier',
          values: { staging: 'warn' },
        },
      },
    )
    expect(accent?.color).toBe('warn')
    expect(accent?.labelKey).toBe('tier')
  })
})
