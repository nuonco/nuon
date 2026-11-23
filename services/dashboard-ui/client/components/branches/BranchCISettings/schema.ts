import { z } from 'zod'
import type { TAppBranchRunConfig } from '@/types'

export const parsePathPatterns = (text: string): string[] =>
  text
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)

export const branchCISettingsSchema = z
  .object({
    ignoreChangesRegex: z.string(),
    sendStatusesOnIgnore: z.boolean(),
    allowedPaths: z.string(),
    deniedPaths: z.string(),
  })
  .refine(
    (values) =>
      !(
        parsePathPatterns(values.allowedPaths).length &&
        parsePathPatterns(values.deniedPaths).length
      ),
    {
      message:
        'Allowed paths and denied paths cannot both be set. Clear one of them.',
      path: ['deniedPaths'],
    }
  )

export type BranchCISettingsValues = z.infer<typeof branchCISettingsSchema>

export type TPathFilter = Pick<
  BranchCISettingsValues,
  'allowedPaths' | 'deniedPaths'
>

export const pathFilterFromRunConfig = (
  runConfig?: TAppBranchRunConfig
): TPathFilter => ({
  allowedPaths: (runConfig?.paths ?? []).join('\n'),
  deniedPaths: (runConfig?.paths_ignore ?? []).join('\n'),
})

export const runConfigWithPathFilter = (
  runConfig: TAppBranchRunConfig | undefined,
  values: TPathFilter
): TAppBranchRunConfig => ({
  ...runConfig,
  paths: parsePathPatterns(values.allowedPaths),
  paths_ignore: parsePathPatterns(values.deniedPaths),
})
