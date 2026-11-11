import { describe, expect, test } from 'bun:test'
import {
  canonicalRepositories,
  parseRepositoryNames,
  repositoryNamesSchema,
} from './repository-scope'

describe('friendly repository scopes', () => {
  test('trims, omits empty lines, and deduplicates without changing paths', () => {
    expect(
      parseRepositoryNames(' acme/api \r\n\nacme/platform/worker\nacme/api')
    ).toEqual(['acme/api', 'acme/platform/worker'])
  })
  test('canonicalizes each cloud scope without tags, image names, or widening', () => {
    expect(
      canonicalRepositories(
        {
          preset: 'ecr',
          target_id: '123456789012',
          default_region: 'us-west-2',
        },
        ['acme/api']
      )
    ).toEqual(['arn:aws:ecr:us-west-2:123456789012:repository/acme/api'])
    expect(
      canonicalRepositories(
        { preset: 'gar', target_id: 'acme-project', default_region: 'us' },
        ['acme-images']
      )
    ).toEqual(['projects/acme-project/locations/us/repositories/acme-images'])
    expect(
      canonicalRepositories({ preset: 'acr', target_id: 'acme-subscription' }, [
        'acme/api',
      ])
    ).toEqual(['acme/api'])
  })
  for (const preset of ['ecr', 'gar', 'acr'] as const) {
    test(`${preset} rejects empty, wildcard, tag, URL, and ARN inputs`, () => {
      for (const value of [
        '',
        ' ',
        'acme/*',
        'acme:latest',
        'https://example.com/acme',
        'arn:aws:ecr:us-west-2:123456789012:repository/acme',
      ]) {
        expect(repositoryNamesSchema(preset).safeParse(value).success).toBe(
          false
        )
      }
    })
  }
  test('keeps valid nested ECR and ACR separators and GAR IDs', () => {
    expect(
      repositoryNamesSchema('ecr').safeParse(
        'acme/api--worker\nacme/api__worker'
      ).success
    ).toBe(true)
    expect(
      repositoryNamesSchema('acr').safeParse('acme/platform/api--worker')
        .success
    ).toBe(true)
    expect(
      repositoryNamesSchema('gar').safeParse('acme-images\nacme_services')
        .success
    ).toBe(true)
    expect(
      repositoryNamesSchema('gar').safeParse('acme-images/api').success
    ).toBe(false)
    expect(
      repositoryNamesSchema('gar').safeParse(
        'projects/acme/locations/us/repositories/acme'
      ).success
    ).toBe(false)
    expect(
      repositoryNamesSchema('ecr').safeParse(
        'projects/acme/locations/us/repositories/acme'
      ).success
    ).toBe(true)
  })
})
