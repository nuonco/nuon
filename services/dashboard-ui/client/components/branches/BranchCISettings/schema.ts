import { z } from 'zod'
import type { TAppBranchRunConfig } from '@/types'

export const MAX_PATH_PATTERNS = 100

export const parsePathPatterns = (text: string): string[] =>
  text
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)

export const branchCISettingsSchema = z.object({
  ignoreChangesRegex: z.string(),
  sendStatusesOnIgnore: z.boolean(),
  pathFilterMode: z.enum(['paths', 'paths_ignore']),
  pathPatterns: z
    .string()
    .refine((text) => parsePathPatterns(text).length <= MAX_PATH_PATTERNS, {
      message: `Use at most ${MAX_PATH_PATTERNS} patterns`,
    }),
})

export type BranchCISettingsValues = z.infer<typeof branchCISettingsSchema>

export const pathFilterFromRunConfig = (
  runConfig?: TAppBranchRunConfig
): Pick<BranchCISettingsValues, 'pathFilterMode' | 'pathPatterns'> =>
  runConfig?.paths_ignore?.length
    ? {
        pathFilterMode: 'paths_ignore',
        pathPatterns: runConfig.paths_ignore.join('\n'),
      }
    : {
        pathFilterMode: 'paths',
        pathPatterns: (runConfig?.paths ?? []).join('\n'),
      }

export const runConfigWithPathFilter = (
  runConfig: TAppBranchRunConfig | undefined,
  values: Pick<BranchCISettingsValues, 'pathFilterMode' | 'pathPatterns'>
): TAppBranchRunConfig => {
  const patterns = parsePathPatterns(values.pathPatterns)
  return {
    ...runConfig,
    paths: values.pathFilterMode === 'paths' ? patterns : [],
    paths_ignore: values.pathFilterMode === 'paths_ignore' ? patterns : [],
  }
}
