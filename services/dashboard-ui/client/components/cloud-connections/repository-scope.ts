import { z } from 'zod'
import type { TCloudConnection } from '@/types'

export const parseRepositoryNames = (value: string) => [
  ...new Set(
    value
      .split(/\r?\n/)
      .map((name) => name.trim())
      .filter(Boolean)
  ),
]

export const repositoryNamesSchema = (preset: string) =>
  z.string().superRefine((value, ctx) => {
    const names = parseRepositoryNames(value)
    const path =
      /^[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*(?:\/[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*)*$/
    const valid =
      names.length > 0 &&
      names.every((name) =>
        preset === 'gar'
          ? /^[a-z][a-z0-9_-]{0,62}$/.test(name)
          : name.length >= (preset === 'ecr' ? 2 : 1) &&
            name.length <= 256 &&
            path.test(name)
      )
    if (!valid)
      ctx.addIssue({
        code: 'custom',
        message:
          preset === 'gar'
            ? 'Enter repository IDs of up to 63 lowercase letters, numbers, underscores, or hyphens, starting with a letter; omit image paths and tags'
            : preset === 'ecr'
              ? 'Enter lowercase repository names (2–256 characters), including any slash-separated path; omit ARNs, registry hosts, wildcards, tags, and digests'
              : 'Enter exact lowercase repository paths, without wildcards, registry hosts, tags, or digests',
      })
  })

export const isImagePreset = (
  preset: TCloudConnection['preset'] | string | undefined
): preset is 'ecr' | 'gar' | 'acr' =>
  preset === 'ecr' || preset === 'gar' || preset === 'acr'

export const canonicalRepositories = (
  {
    preset,
    target_id,
    default_region,
  }: Pick<TCloudConnection, 'preset' | 'target_id' | 'default_region'>,
  names: string[]
) =>
  names.map((name) =>
    preset === 'ecr'
      ? `arn:aws:ecr:${default_region}:${target_id}:repository/${name}`
      : preset === 'gar'
        ? `projects/${target_id}/locations/${default_region}/repositories/${name}`
        : name
  )
