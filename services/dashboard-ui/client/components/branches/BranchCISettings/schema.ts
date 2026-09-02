import { z } from 'zod'

export const branchCISettingsSchema = z.object({
  ignoreChangesRegex: z.string(),
  sendStatusesOnIgnore: z.boolean(),
})

export type BranchCISettingsValues = z.infer<typeof branchCISettingsSchema>
